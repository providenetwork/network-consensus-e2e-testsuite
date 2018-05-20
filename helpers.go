package main

func stringOrNil(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}
