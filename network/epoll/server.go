/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package epoll

import (
	"io"
	"net"
	"time"

	"go.osspkg.com/x/errors"
	"go.osspkg.com/x/logx"
	"go.osspkg.com/x/network/address"
	"go.osspkg.com/x/syncing"
	"go.osspkg.com/x/xc"
	"golang.org/x/sys/unix"
)

type (
	Config struct {
		Addr            string        `yaml:"addr"`
		ReadTimeout     time.Duration `yaml:"read_timeout,omitempty"`
		WriteTimeout    time.Duration `yaml:"write_timeout,omitempty"`
		IdleTimeout     time.Duration `yaml:"idle_timeout,omitempty"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout,omitempty"`
	}

	Server struct {
		status   syncing.Switch
		wg       syncing.Group
		handler  Handler
		log      logx.Logger
		conf     Config
		eof      []byte
		listener net.Listener
		epoll    *epoll
	}
)

func New(conf Config, handler Handler, eof []byte, l logx.Logger) *Server {
	return &Server{
		status:  syncing.NewSwitch(),
		wg:      syncing.NewGroup(),
		conf:    conf,
		handler: handler,
		log:     l,
		eof:     eof,
	}
}

func (s *Server) validate() {
	s.conf.Addr = address.Check(s.conf.Addr)
	if len(s.eof) == 0 {
		s.eof = defaultEOF
	}
}

func (s *Server) Up(ctx xc.Context) (err error) {
	if !s.status.On() {
		return errServAlreadyRunning
	}
	s.validate()
	if s.listener, err = net.Listen("tcp", s.conf.Addr); err != nil {
		return
	}
	if s.epoll, err = newEpoll(s.log); err != nil {
		return
	}
	s.log.WithFields(logx.Fields{"ip": s.conf.Addr}).Infof("Epoll server started")
	s.wg.Background(func() {
		s.connAccept(ctx)
	})
	s.wg.Background(func() {
		s.epollAccept(ctx)
	})
	return
}

func (s *Server) Down() error {
	if !s.status.Off() {
		return errServAlreadyStopped
	}
	err := errors.Wrap(s.epoll.CloseAll(), s.listener.Close())
	s.wg.Wait()
	if err != nil {
		s.log.WithFields(logx.Fields{
			"err": err.Error(),
			"ip":  s.conf.Addr,
		}).Errorf("Epoll server stopped")
		return err
	}
	s.log.WithFields(logx.Fields{
		"ip": s.conf.Addr,
	}).Infof("Epoll server stopped")
	return nil
}

func (s *Server) connAccept(ctx xc.Context) {
	defer func() {
		ctx.Close()
	}()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				s.log.WithFields(logx.Fields{"err": err.Error()}).Errorf("Epoll conn accept")
				// TODO: check error?
				// var ne net.Error
				// if errors.As(err, ne) {
				//	time.Sleep(1 * time.Second)
				//	continue
				// }
				return
			}
		}
		if err = s.epoll.AddOrClose(conn); err != nil {
			s.log.WithFields(logx.Fields{
				"err": err.Error(), "ip": conn.RemoteAddr().String(),
			}).Errorf("Epoll add conn")
		}
	}
}

func (s *Server) epollAccept(ctx xc.Context) {
	defer func() {
		ctx.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			list, err := s.epoll.Wait()
			switch true {
			case err == nil:
			case errors.Is(err, errEpollEmptyEvents):
				continue
			case errors.Is(err, unix.EINTR):
				continue
			default:
				s.log.WithFields(logx.Fields{
					"err": err.Error(),
				}).Errorf("Epoll accept conn")
				continue
			}

			for _, c := range list {
				c := c
				go func(conn *epollNetItem) {
					defer conn.Await(false)

					if err1 := newEpollConn(conn.Conn, s.handler, s.eof); err1 != nil {
						if err2 := s.epoll.Close(conn); err2 != nil {
							s.log.WithFields(logx.Fields{
								"err": err2.Error(),
								"ip":  conn.Conn.RemoteAddr().String(),
							}).Errorf("Epoll add conn")
						}
						if errors.Is(err1, io.EOF) {
							s.log.WithFields(logx.Fields{
								"err": err1.Error(),
								"ip":  conn.Conn.RemoteAddr().String(),
							}).Errorf("Epoll bad conn")
						}
					}
				}(c)
			}
		}
	}
}