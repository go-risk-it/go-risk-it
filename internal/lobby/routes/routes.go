package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	lobbyRequest "github.com/go-risk-it/go-risk-it/internal/lobby/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/rest/response"
	lobbyWs "github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
)

func ProvideRoutes(
	creationCtrl *CreationController,
	managementCtrl *ManagementController,
	startCtrl *StartController,
	lobbyConnectionManager lobbyWs.Manager,
	upgrader ws.Upgrader,
) []*route.Route {
	return []*route.Route{
		route.Authed("POST /api/v1/lobbies", createLobby(creationCtrl)),
		route.Authed("GET /api/v1/lobbies/summary", getLobbiesSummary(managementCtrl)),
		route.Lobby("POST /api/v1/lobbies/{id}/join", joinLobby(managementCtrl)),
		route.Lobby("POST /api/v1/lobbies/{id}/start", startGame(startCtrl)),
		route.LobbyWS(
			"GET /api/v1/lobbies/{id}/ws",
			connectLobbyWS(lobbyConnectionManager, upgrader),
		),
	}
}

func createLobby(creationCtrl *CreationController) route.PlainHandler {
	return func(writer http.ResponseWriter, req *http.Request) error {
		createLobbyRequest, err := restutils.DecodeRequest[lobbyRequest.CreateLobby](writer, req)
		if err != nil {
			return err
		}

		userContext, ok := req.Context().(ctx.UserContext)
		if !ok {
			return errors.New("invalid user context")
		}

		lobbyID, err := creationCtrl.CreateLobby(userContext, createLobbyRequest)
		if err != nil {
			return err
		}

		resp, err := json.Marshal(response.CreateLobby{LobbyID: lobbyID})
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		restutils.WriteResponse(writer, resp, http.StatusCreated)

		return nil
	}
}

func getLobbiesSummary(managementCtrl *ManagementController) route.PlainHandler {
	return func(writer http.ResponseWriter, req *http.Request) error {
		userContext, ok := req.Context().(ctx.UserContext)
		if !ok {
			return errors.New("invalid user context")
		}

		lobbies, err := managementCtrl.GetUserLobbies(userContext)
		if err != nil {
			return err
		}

		lobbiesResponse, err := json.Marshal(lobbies)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		restutils.WriteResponse(writer, lobbiesResponse, http.StatusOK)

		return nil
	}
}

func joinLobby(managementCtrl *ManagementController) route.LobbyHandler {
	return func(writer http.ResponseWriter, req *http.Request, lobbyCtx ctx.LobbyContext) error {
		joinLobbyRequest, err := restutils.DecodeRequest[lobbyRequest.JoinLobby](writer, req)
		if err != nil {
			return err
		}

		if err := managementCtrl.JoinLobby(lobbyCtx, joinLobbyRequest); err != nil {
			return err
		}

		writer.WriteHeader(http.StatusNoContent)

		return nil
	}
}

func startGame(startCtrl *StartController) route.LobbyHandler {
	return func(_ http.ResponseWriter, _ *http.Request, lc ctx.LobbyContext) error {
		return startCtrl.StartGame(lc)
	}
}

func connectLobbyWS(
	lobbyConnectionManager lobbyWs.Manager,
	upgrader ws.Upgrader,
) route.LobbyHandler {
	return func(
		writer http.ResponseWriter,
		request *http.Request,
		lobbyCtx ctx.LobbyContext,
	) error {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return fmt.Errorf("unable to upgrade websocket connection: %w", err)
		}

		lobbyConnectionManager.ConnectPlayer(lobbyCtx, conn)
		slog.InfoContext(
			request.Context(),
			"Lobby WS upgraded",
			"remoteAddress",
			conn.RemoteAddr().String(),
		)

		return nil
	}
}
