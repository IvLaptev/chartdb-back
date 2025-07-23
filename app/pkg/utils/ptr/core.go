package ptr

func To[G any](v G) *G {
	return &v
}

func From[G any](v *G) G {
	return *v
}
