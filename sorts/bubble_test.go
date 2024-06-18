/*
 *  Copyright (c) 2019-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sorts_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.osspkg.com/algorithms/sorts"
)

func TestUnit_SortBubble(t *testing.T) {
	tests := []struct {
		name string
		args []int
		want []int
	}{
		{
			name: "IntCase1",
			args: nil,
			want: nil,
		},
		{
			name: "IntCase2",
			args: []int{1, 67, 23, 1, 5, 9, 5, 32, 1, 34, 68, 9, 5, 23, 0, 0, 0, 0, 0, 5, 5, 3, 2, 1},
			want: []int{0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 3, 5, 5, 5, 5, 5, 9, 9, 23, 23, 32, 34, 67, 68},
		},
		{
			name: "IntCase3",
			args: []int{1},
			want: []int{1},
		},
		{
			name: "IntCase4",
			args: []int{9, 4, 1, 6, 0},
			want: []int{0, 1, 4, 6, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorts.Bubble(tt.args, func(i, j int) bool {
				return tt.args[i] < tt.args[j]
			})
			require.Equal(t, tt.want, tt.args)
		})
	}
}

func Benchmark_SortBubble(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		arr := []int{30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
		sorts.Bubble(arr, func(i, j int) bool {
			return arr[i] < arr[j]
		})
	}
}