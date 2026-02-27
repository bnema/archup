package assets

import _ "embed"

//go:embed logo.txt
var LogoASCII string

//go:embed arch-logo.png
var LimineLogo []byte
