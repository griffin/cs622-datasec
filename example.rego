package sql

default allow = false

contains(arr, elem) {
    arr[_] = elem
}

allow {
    not contains(input.cols, "bank_id")
    input.star = false
}
