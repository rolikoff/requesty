package main

// The entry point for the app.
func main() {
	a := App{}
	defer a.Dispose()
	a.Initialize("./requesty.db")
	a.Run(":8080")
}
