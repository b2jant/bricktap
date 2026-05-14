package main

import (
	"fmt"
	"github.com/b2jant/bricktap/internal/core"
	"github.com/b2jant/bricktap/internal/parser"
)

func main() {
	model, err := parser.ParseModel("semantic_models/sales/cur/cur_sales_orders.yaml", core.GlobalRules{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", model.BaseEntity)
}
