package sel

import "sel/b"

var X = b.C.g // error, not exported, disregard ambiguity
