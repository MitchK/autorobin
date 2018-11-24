package quandl

type response struct {
	DatasetData datasetData `json:"dataset_data,omitempty"`
	QuandlError quandlError `json:"quandl_error,omitempty"`
}

type quandlError struct {
	Message string `json:"message"`
}
type datasetData struct {
	ColumnNames []string        `json:"column_names"`
	Order       string          `json:"order"`
	Data        [][]interface{} `json:"data"`
}
