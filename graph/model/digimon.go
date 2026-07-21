package model

type Digimon struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Level                 string   `json:"level"`
	Type                  string   `json:"type"`
	Attribute             string   `json:"attribute"`
	Moves                 []string `json:"moves"`
	ImgSrc                string   `json:"img_src"`
	Background            string   `json:"background"`
	PreviousDigivolutions []string `json:"previous_digivolutions"`
	NextDigivolutions     []string `json:"next_digivolutions"`
	IsMode                bool     `json:"is_mode"`
	IsXAntibody           bool     `json:"is_x_antibody"`
}
