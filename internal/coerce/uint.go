package coerce

func ToUint64(v interface{}) uint64 {
	return uint64(ToInt64(v))
}

