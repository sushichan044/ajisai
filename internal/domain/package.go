package domain

import "encoding/xml"

type (
	AgentPresetPackage struct {
		PackageName string         `xml:"name,attr"`
		Presets     []*AgentPreset `xml:"preset,omitempty"`
	}
)

// XML marshalling implementation

type (
	xmlPackage struct {
		XMLName     xml.Name     `xml:"package"`
		PackageName string       `xml:"name,attr"`
		Presets     []*xmlPreset `xml:"preset,omitempty"`
	}
)

func (p *AgentPresetPackage) MarshalToXML() ([]byte, error) {
	return p.toXML().Marshal()
}

func (p *AgentPresetPackage) toXML() *xmlPackage {
	outputPkg := xmlPackage{
		PackageName: p.PackageName,
	}

	if len(p.Presets) > 0 {
		outputPkg.Presets = make([]*xmlPreset, len(p.Presets))
		for i, preset := range p.Presets {
			outputPkg.Presets[i] = preset.toXML()
		}
	}

	return &outputPkg
}

func (xp *xmlPackage) Marshal() ([]byte, error) {
	return xml.MarshalIndent(xp, "", "  ")
}
