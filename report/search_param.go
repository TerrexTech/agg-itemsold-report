package report

type Comparator struct {
	Lt float64     `json:"$lt,omitempty"`
	Gt float64     `json:"$gt,omitempty"`
	Eq interface{} `json:"$eq,omitempty"`
}

// {
// 	sku: {
// 		$eq: "trestsda",
// 	}
// 	name: {
// 		$eq: "namdkgsfjsdio",
// 	}
// 	timestamp: {
// 		$lt: 5,
// 		$gt, 2,

// 		$eq: 234654,
// 	}
// }
