package main

import (
    "slices"
)

func computeSmithWaterman(a, b string) int {
    n := len(a)+1
    m := len(b)+1
    stride := m
    table := make([]int, n*m)

    for i := 1; i < n; i++ {
        for j := 1; j < m; j++ {
            var s int
            if a[i-1] != b[j-1] {
                s = -1
            } else {
                s = 2
            }
            table[i*stride+j] = max(0, table[(i-1)*stride+(j-1)]+s, table[(i-1)*stride + j] + s, table[i*stride+(j-1)]+s)
        }
    }
    return slices.Max(table)
}
