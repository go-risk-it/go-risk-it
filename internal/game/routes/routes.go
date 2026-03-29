package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	gameRequest "github.com/go-risk-it/go-risk-it/internal/game/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/response"
	gameWs "github.com/go-risk-it/go-risk-it/internal/game/ws"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
)

func ProvideRoutes(
	gameCtrl *GameController,
	advCtrl *AdvancementController,
	moveCtrl *MoveController,
	gameConnectionManager gameWs.Manager,
	upgrader ws.Upgrader,
) []*route.Route {
	return []*route.Route{
		route.Authed("POST /api/v1/games", createGame(gameCtrl)),
		route.Authed("GET /api/v1/games/summary", getGamesSummary(gameCtrl)),
		route.Game("POST /api/v1/games/{id}/advancements", advanceGame(advCtrl)),
		route.Game(
			"POST /api/v1/games/{id}/moves/deployments",
			moveHandler[gameRequest.DeployMove](moveCtrl.PerformDeployMove),
		),
		route.Game(
			"POST /api/v1/games/{id}/moves/attacks",
			moveHandler[gameRequest.AttackMove](moveCtrl.PerformAttackMove),
		),
		route.Game(
			"POST /api/v1/games/{id}/moves/conquers",
			moveHandler[gameRequest.ConquerMove](moveCtrl.PerformConquerMove),
		),
		route.Game(
			"POST /api/v1/games/{id}/moves/reinforcements",
			moveHandler[gameRequest.ReinforceMove](moveCtrl.PerformReinforceMove),
		),
		route.Game(
			"POST /api/v1/games/{id}/moves/cards",
			moveHandler[gameRequest.CardsMove](moveCtrl.PerformCardsMove),
		),
		route.GameWS("GET /api/v1/games/{id}/ws", connectGameWS(gameConnectionManager, upgrader)),
	}
}

func createGame(gameCtrl *GameController) route.PlainHandler {
	return func(writer http.ResponseWriter, req *http.Request) error {
		createGameRequest, err := restutils.DecodeRequest[gameRequest.CreateGame](writer, req)
		if err != nil {
			return err
		}

		userContext, ok := req.Context().(ctx.UserContext)
		if !ok {
			return errors.New("invalid user context")
		}

		gameID, err := gameCtrl.CreateGame(userContext, createGameRequest) //nolint:contextcheck
		if err != nil {
			return err
		}

		createGameResponse, err := json.Marshal(response.CreateGame{GameID: gameID})
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		restutils.WriteResponse(writer, createGameResponse, http.StatusCreated)

		return nil
	}
}

func getGamesSummary(gameCtrl *GameController) route.PlainHandler {
	return func(writer http.ResponseWriter, req *http.Request) error {
		userContext, ok := req.Context().(ctx.UserContext)
		if !ok {
			return errors.New("invalid user context")
		}

		games, err := gameCtrl.GetUserGames(userContext)
		if err != nil {
			return err
		}

		gamesResponse, err := json.Marshal(games)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		restutils.WriteResponse(writer, gamesResponse, http.StatusOK)

		return nil
	}
}

func advanceGame(advCtrl *AdvancementController) route.GameHandler {
	return func(writer http.ResponseWriter, req *http.Request, gameCtx ctx.GameContext) error {
		advancementRequest, err := restutils.DecodeRequest[gameRequest.Advancement](writer, req)
		if err != nil {
			return err
		}

		if err = advCtrl.Advance(gameCtx, advancementRequest); err != nil {
			return err
		}

		restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)

		return nil
	}
}

func moveHandler[T any](
	perform func(ctx.GameContext, T) error,
) route.GameHandler {
	return func(writer http.ResponseWriter, req *http.Request, gameCtx ctx.GameContext) error {
		moveRequest, err := restutils.DecodeRequest[T](writer, req)
		if err != nil {
			return err
		}

		if err := perform(gameCtx, moveRequest); err != nil {
			return err
		}

		restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)

		return nil
	}
}

func connectGameWS(
	gameConnectionManager gameWs.Manager,
	upgrader ws.Upgrader,
) route.GameHandler {
	return func(writer http.ResponseWriter, request *http.Request, gameCtx ctx.GameContext) error {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return fmt.Errorf("unable to upgrade websocket connection: %w", err)
		}

		gameConnectionManager.ConnectPlayer(gameCtx, conn)
		slog.InfoContext(
			request.Context(),
			"Game WS upgraded",
			"remoteAddress",
			conn.RemoteAddr().String(),
		)

		return nil
	}
}
