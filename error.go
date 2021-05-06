package shopifyoauth

import "errors"

var ErrURLVerification = errors.New("url verification failed")
var ErrStateVerification = errors.New("value of OAuth state is not matched")
