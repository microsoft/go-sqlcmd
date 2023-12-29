package mechanism

var errorCallback func(err error)

func checkErr(err error) {
	errorCallback(err)
}
