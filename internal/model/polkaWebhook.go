package model

type PolkaWebbHook struct {
	Event string `json:"event"`
	Data  struct {
		User_id int `json:"user_id"`
	} `json:"data"`
}
