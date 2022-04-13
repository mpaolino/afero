package badfs

//type errorTrigger struct {
//	start int64
//	end   int64
//}

type RandomError struct {
	err         error
	probability float64
	//	errorTrigger
}

func NewRandomError(err error, probability float64) *RandomError {
	return &RandomError{err: err, probability: probability}
}

func (r *RandomError) getError() error {
	return r.err
}
