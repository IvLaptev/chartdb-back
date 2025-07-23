package storage

type RequestOptions struct {
	UseLock bool
}

func NewOptions(opts []RequestOption) RequestOptions {
	options := RequestOptions{
		UseLock: false,
	}

	for _, o := range opts {
		o(&options)
	}
	return options
}

type RequestOption func(*RequestOptions)

func WithLock() RequestOption {
	return func(opts *RequestOptions) {
		opts.UseLock = true
	}
}
