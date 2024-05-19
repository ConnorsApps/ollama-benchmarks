package main

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func ptr[K any](a K) *K {
	return &a
}
