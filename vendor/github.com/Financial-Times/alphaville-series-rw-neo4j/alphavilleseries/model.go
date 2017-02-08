package alphavilleseries

type AlphavilleSeries struct {
	UUID                   string                 `json:"uuid"`
	PrefLabel              string                 `json:"prefLabel"`
	AlternativeIdentifiers alternativeIdentifiers `json:"alternativeIdentifiers"`
	Types                  []string               `json:"types,omitempty"`
}

type alternativeIdentifiers struct {
	TME   []string `json:"TME,omitempty"`
	UUIDS []string `json:"uuids"`
}

const (
	tmeIdentifierLabel = "TMEIdentifier"
	uppIdentifierLabel = "UPPIdentifier"
)

type AlphavilleSeriesLink struct {
	ApiUrl string `json:"apiUrl"`
}
