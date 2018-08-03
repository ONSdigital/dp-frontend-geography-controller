package mapper

import (
	"os"

	"github.com/ONSdigital/dp-frontend-models/model"
)

// SetTaxonomyDomain will set the taxonomy domain for a given pages
func SetTaxonomyDomain(p *model.Page) {
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")
}
