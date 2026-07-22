package model

type Digimon struct {
	ID                    string   `json:"_id" bson:"_id,omitempty"`
	Name                  string   `json:"name" bson:"name,omitempty"`
	Level                 string   `json:"level" bson:"level,omitempty"`
	Type                  string   `json:"type" bson:"type,omitempty"`
	Attribute             string   `json:"attribute" bson:"attribute,omitempty"`
	Moves                 []string `json:"moves" bson:"moves,omitempty"`
	ImgSrc                string   `json:"img_src" bson:"img_src,omitempty"`
	Background            string   `json:"background" bson:"background,omitempty"`
	PreviousDigivolutions []string `json:"previous_digivolutions" bson:"previous_digivolutions,omitempty"`
	NextDigivolutions     []string `json:"next_digivolutions" bson:"next_digivolutions,omitempty"`
	IsMode                bool     `json:"is_mode" bson:"is_mode,omitempty"`
	IsXAntibody           bool     `json:"is_x_antibody" bson:"is_x_antibody,omitempty"`
}
