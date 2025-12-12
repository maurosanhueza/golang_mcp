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
		"GithubIssuesMCP",
		"0.0.1",
		server.WithLogging())

	// ahora definimos la herramienta que queremos utilizar
	tool := mcp.NewTool("get_issues",
		mcp.WithDescription("Obtiene las issues abierta de un repositorio"),
		//agregamos los parametros que necesita la herramienta para fucionar
		mcp.WithString("owner", mcp.Required(), mcp.Description("Propietario del repo (org o user)")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Nombre del repositorio")),
	)

	//agragamos la herramienta al server
	s.AddTool(
		tool,
		// funcion handler que ejecuta la herramienta
		// req es el request del MCP Host (Cloud Desktop)
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := req.Params.Arguments.(map[string]any)
			if !ok {
				return mcp.NewToolResultError("Arguments no es un map[string]any"), nil
			}
			// Obtener owner del repo
			owner, ok := args["owner"].(string)
			if !ok {
				return mcp.NewToolResultError("owner debe ser string"), nil
			}
			// Obtener nombre del repo
			repo, ok := args["repo"].(string)
			if !ok {
				return mcp.NewToolResultError("repo debe ser string"), nil
			}

			url := fmt.Sprintf("https://api.golang.com/repos/%s/%s/issues", owner, repo)

			// configuramos llamada a la api de github utilizanod token generado en github
			request, _ := http.NewRequest("GET", url, nil)
			request.Header.Set("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))

			//ejecutamos peticion a la apia de github
			resp, err := http.DefaultClient.Do(request)

			//validamos errores
			if err != nil {
				return mcp.NewToolResultErrorFromErr("falló petición a github", err), nil
			}
			defer resp.Body.Close()

			//controlamos el status
			if resp.StatusCode != http.StatusOK {
				var apiErr struct {
					Message string `json:"message"`
				}

				if decErr := json.NewDecoder(resp.Body).Decode(&apiErr); decErr != nil {
					return mcp.NewToolResultError("error leyendo JSON de error"), nil
				}
				return mcp.NewToolResultError(apiErr.Message), nil
			}

			//decodificamos info que traemos de la api
			var issues []Issue
			if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
				return mcp.NewToolResultErrorFromErr("error decodificando JSON de issues", err), nil
			}

			var contents []mcp.Content
			//agregamos al content url formateada
			for _, i := range issues {
				contents = append(contents, mcp.NewTextContent(
					fmt.Sprintf("%d [%s] %s", i.Number, i.State, i.Title),
				))
			}
			return &mcp.CallToolResult{Content: contents}, nil
		})

	//iniciar el servidor
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Error servidor $v\n", err)
	}
}
