package validator

import "net/url"

func Target(v string) bool { u, e := url.Parse(v); return e == nil && u.Scheme != "" && u.Host != "" }
