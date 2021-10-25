package newrelic

import (
	"bytes"
)

func split(target []byte) ([]byte, []byte) {
	// 58 according to the ASCII table is -> :
	pos := bytes.IndexByte(target, 58)

	// not found
	if pos == -1 {
		return nil, nil
	}

	// ":foo" or ":"
	if pos == 0 {
		return nil, nil
	}

	/*
		we should split headers into 2 parts. Parts are separated by the colon (:)
		"foo:bar"
		we should not panic on cases like:
		":foo"
		"foo: bar"
		:
	*/

	// handle case like this "bar:"
	if len(target) < pos+1 {
		return nil, nil
	}

	return target[:pos], target[pos+1:]
}
