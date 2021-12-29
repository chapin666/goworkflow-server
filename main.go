package main

import (
	"errors"
	"github.com/chapin666/kitten"
	"github.com/chapin666/kitten/model"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

type pageParam struct {
	PageIndex uint `json:"page_index" form:"page_index" query:"page_index"`
	PageSize  uint `json:"page_size" form:"page_size" query:"page_size"`
}

func (p *pageParam) Validate() error {
	if p.PageIndex <= 0 {
		p.PageIndex = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return nil
}

type saveFlowRequest struct {
	Data string `json:"data"`
}

func (c *saveFlowRequest) Validate() error {
	if len(c.Data) == 0 {
		return errors.New("请求含有空数据")
	}
	return nil
}

type flowQueryParam struct {
	pageParam
	Code     string `query:"code"`
	Name     string `query:"name"`
	TypeCode string `query:"type_code"`
	Status   int    `query:"status"`
}

func saveFlow(engine *kitten.Engine) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var req saveFlowRequest
		if err := ctx.Bind(&req); err != nil {
			return err
		}
		if err := req.Validate(); err != nil {
			return err
		}

		recordID, err := engine.SaveFlow([]byte(req.Data))
		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, recordID)
	}
}

func flowList(engine *kitten.Engine) echo.HandlerFunc {
	return func(ctx echo.Context) error {

		query := new(flowQueryParam)
		if err := ctx.Bind(query); err != nil {
			return err
		}
		query.Validate()

		pageIndex := query.PageIndex
		pageSize := query.PageSize
		params := model.FlowQueryParam{
			Code:     query.Code,
			Name:     query.Name,
			TypeCode: query.TypeCode,
			Status:   query.Status,
		}
		total, items, err := engine.QueryAllFlowPage(params, pageIndex, pageSize)
		if err != nil {
			return err
		}
		response := map[string]interface{}{
			"list": items,
			"pagination": map[string]interface{}{
				"total":    total,
				"current":  pageIndex,
				"pageSize": pageSize,
			},
		}
		return ctx.JSON(http.StatusOK, response)
	}
}

func getFlow(engine *kitten.Engine) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		item, err := engine.GetFlow(ctx.Param("id"))
		if err != nil {
			return err
		}
		return ctx.JSON(http.StatusOK, item)
	}
}

func delFlow(engine *kitten.Engine) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		err := engine.DeleteFlow(ctx.Param("id"))
		if err != nil {
			return err
		}
		return ctx.JSON(http.StatusOK, "ok")
	}
}

func main() {
	workflowEngine, err := kitten.New("root@tcp(127.0.0.1:3306)/flow_test?charset=utf8", true)
	if err != nil {
		panic(err)
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/flow", saveFlow(workflowEngine))
	e.GET("/flow/page", flowList(workflowEngine))
	e.GET("/flow/:id", getFlow(workflowEngine))
	e.DELETE("/flow/:id", delFlow(workflowEngine))

	// Start server
	e.Logger.Fatal(e.Start(":6062"))
}
