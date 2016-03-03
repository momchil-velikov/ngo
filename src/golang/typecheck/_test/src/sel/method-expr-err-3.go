package sel

import "sel/a"

var X = a.A.g // error, not exported
