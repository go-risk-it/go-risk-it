package request_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/lobby/api/rest/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLobby_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     request.CreateLobby
		wantErr string
	}{
		{name: "valid", req: request.CreateLobby{OwnerName: "Giovanni"}},
		{name: "empty", req: request.CreateLobby{OwnerName: ""}, wantErr: "ownerName"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.req.Validate()
			if testCase.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
			}
		})
	}
}

func TestJoinLobby_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     request.JoinLobby
		wantErr string
	}{
		{name: "valid", req: request.JoinLobby{ParticipantName: "Giovanni"}},
		{name: "empty", req: request.JoinLobby{ParticipantName: ""}, wantErr: "participantName"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.req.Validate()
			if testCase.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
			}
		})
	}
}
