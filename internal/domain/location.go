package domain

// KenyanCounties is the canonical list of Kenya's 47 counties.
// Profile locations are validated against this list (see the `kenyacounty`
// rule in pkg/validator). Keep frontend dropdowns in sync with these strings.
var KenyanCounties = []string{
	"Mombasa",
	"Kwale",
	"Kilifi",
	"Tana River",
	"Lamu",
	"Taita Taveta",
	"Garissa",
	"Wajir",
	"Mandera",
	"Marsabit",
	"Isiolo",
	"Meru",
	"Tharaka Nithi",
	"Embu",
	"Kitui",
	"Machakos",
	"Makueni",
	"Nyandarua",
	"Nyeri",
	"Kirinyaga",
	"Muranga",
	"Kiambu",
	"Turkana",
	"West Pokot",
	"Samburu",
	"Trans Nzoia",
	"Uasin Gishu",
	"Elgeyo Marakwet",
	"Nandi",
	"Baringo",
	"Laikipia",
	"Nakuru",
	"Narok",
	"Kajiado",
	"Kericho",
	"Bomet",
	"Kakamega",
	"Vihiga",
	"Bungoma",
	"Busia",
	"Siaya",
	"Kisumu",
	"Homa Bay",
	"Migori",
	"Kisii",
	"Nyamira",
	"Nairobi",
}

// IsKenyanCounty reports whether the given string is a recognised county.
func IsKenyanCounty(county string) bool {
	for _, c := range KenyanCounties {
		if c == county {
			return true
		}
	}
	return false
}
