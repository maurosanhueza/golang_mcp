package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Issue struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"htmlurl"`
}

func main() {
	// creamos uun server
	s := server.NewMCPServer(
		"Githubissues",
		"0.0.1",
		server.WithLogging())

	// ahora definimos la herramienta que queremos utilizar
	tool := mcp.NewTool("get_issues",
		mcp.WithDescription("Obtiene las issues abierta de un repositorio"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Propietario del repo (org o user)")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Nombre del repositorio")),
	)

	s.AddTool(
		tool,
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := req.Params.Arguments.(map[string]any)
			if !ok {
				return mcp.NewToolResultError("Arguments no es un map[string]any"), nil
			}

			// 2️⃣ Obtener los parámetros del mapa
			owner, ok := args["owner"].(string)
			if !ok {
				return mcp.NewToolResultError("owner debe ser string"), nil
			}

			repo, ok := args["repo"].(string)
			if !ok {
				return mcp.NewToolResultError("repo debe ser string"), nil
			}

			url := fmt.Sprintf("https://api.golang.com/repos/%s/%s/issues", owner, repo)

			request, _ := http.NewRequest("GET", url, nil)
			request.Header.Set("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))

			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("falló peticion a github", err), nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				var apiErr struct {
					Message string `json:"message"`
				}

				if decErr := json.NewDecoder(resp.Body).Decode(&apiErr); decErr != nil {
					return mcp.NewToolResultError("error leyendo JSNO de error"), nil
				}
				return mcp.NewToolResultError(apiErr.Message), nil

			}
		})
}
