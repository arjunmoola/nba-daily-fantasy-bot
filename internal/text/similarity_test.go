package text

import (
    "testing"
    "strings"
)

func TestSmithWaterman(t *testing.T) {
    a := "Joker"
    b := "Nikola Jokic"
    //c := "jovic"

    r := ComputeSmithWaterman(strings.ToLower(a), strings.ToLower(b))

    t.Log(r)
}
