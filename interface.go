package iex

type Symbol struct {
	// Refers to the symbol represented in Nasdaq Integrated symbology (INET).
	Ticker string
	// Refers to the name of the company or security.
	Name string
	// Refers to the date the symbol reference data was generated.
	Date string
	// Will be true if the symbol is enabled for trading on IEX.
	IsEnabled bool
}
