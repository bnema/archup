package phases

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces/mocks"
	"github.com/bnema/archup/internal/logger"
	"go.uber.org/mock/gomock"
)

// TestBootstrapPhasePreCheck tests internet connectivity validation
func TestBootstrapPhasePreCheck(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*mocks.MockHTTPClient)
		wantErr    bool
		errContains string
	}{
		{
			name: "Internet connectivity - success",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}
				mockHTTP.EXPECT().Get("https://raw.githubusercontent.com").Return(resp, nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Internet connectivity - connection failed",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient) {
				mockHTTP.EXPECT().Get("https://raw.githubusercontent.com").Return(nil, fmt.Errorf("connection refused")).Times(1)
			},
			wantErr:     true,
			errContains: "no internet connectivity",
		},
		{
			name: "Internet connectivity - timeout",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient) {
				mockHTTP.EXPECT().Get("https://raw.githubusercontent.com").Return(nil, fmt.Errorf("context deadline exceeded")).Times(1)
			},
			wantErr:     true,
			errContains: "no internet connectivity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTP := mocks.NewMockHTTPClient(ctrl)
			mockFS := mocks.NewMockFileSystem(ctrl)

			tt.setupMocks(mockHTTP)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewBootstrapPhase(cfg, log, mockHTTP, mockFS)

			err = phase.PreCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("PreCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if errMsg := err.Error(); tt.errContains != "" && !containsSubstring(errMsg, tt.errContains) {
					t.Errorf("PreCheck() error = %q, should contain %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

// TestBootstrapPhaseDownloadFile tests file download functionality
func TestBootstrapPhaseDownloadFile(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		setupMocks  func(*mocks.MockHTTPClient, *mocks.MockFileSystem)
		wantErr     bool
		errContains string
	}{
		{
			name: "Download file - success",
			url:  "https://example.com/file.txt",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient, mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString("file content")),
				}
				mockHTTP.EXPECT().Get("https://example.com/file.txt").Return(resp, nil).Times(1)

				mockFile := NewMockWriteCloser()
				mockFS.EXPECT().Create(gomock.Any()).Return(mockFile, nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Download file - HTTP 404 not found",
			url:  "https://example.com/missing.txt",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient, mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Status:     "404 Not Found",
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}
				mockHTTP.EXPECT().Get("https://example.com/missing.txt").Return(resp, nil).Times(1)
			},
			wantErr:     true,
			errContains: "HTTP 404",
		},
		{
			name: "Download file - HTTP 500 server error",
			url:  "https://example.com/error.txt",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient, mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}
				mockHTTP.EXPECT().Get("https://example.com/error.txt").Return(resp, nil).Times(1)
			},
			wantErr:     true,
			errContains: "HTTP 500",
		},
		{
			name: "Download file - HTTP request failed",
			url:  "https://example.com/file.txt",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient, mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				mockHTTP.EXPECT().Get("https://example.com/file.txt").Return(nil, fmt.Errorf("connection timeout")).Times(1)
			},
			wantErr:     true,
			errContains: "HTTP request failed",
		},
		{
			name: "Download file - directory creation failed",
			url:  "https://example.com/file.txt",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient, mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(fmt.Errorf("permission denied")).Times(1)
			},
			wantErr:     true,
			errContains: "failed to create directory",
		},
		{
			name: "Download file - file creation failed",
			url:  "https://example.com/file.txt",
			setupMocks: func(mockHTTP *mocks.MockHTTPClient, mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString("content")),
				}
				mockHTTP.EXPECT().Get("https://example.com/file.txt").Return(resp, nil).Times(1)

				mockFS.EXPECT().Create(gomock.Any()).Return(nil, fmt.Errorf("no space left")).Times(1)
			},
			wantErr:     true,
			errContains: "failed to create file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTP := mocks.NewMockHTTPClient(ctrl)
			mockFS := mocks.NewMockFileSystem(ctrl)

			tt.setupMocks(mockHTTP, mockFS)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewBootstrapPhase(cfg, log, mockHTTP, mockFS)

			err = phase.downloadFile(tt.url, filepath.Join(tmpDir, "dest.txt"))
			if (err != nil) != tt.wantErr {
				t.Errorf("downloadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if errMsg := err.Error(); tt.errContains != "" && !containsSubstring(errMsg, tt.errContains) {
					t.Errorf("downloadFile() error = %q, should contain %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

// TestBootstrapPhaseExecute tests the Execute method
func TestBootstrapPhaseExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockFS := mocks.NewMockFileSystem(ctrl)

	// Setup successful downloads
	mockFS.EXPECT().MkdirAll(config.DefaultInstallDir, gomock.Any()).Return(nil).Times(1)

	// Expect MkdirAll calls for each file's directory
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Expect Create calls for each file
	mockFile := NewMockWriteCloser()
	mockFS.EXPECT().Create(gomock.Any()).Return(mockFile, nil).AnyTimes()

	// Expect HTTP Get calls for each file
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(bytes.NewBufferString("file content")),
	}
	mockHTTP.EXPECT().Get(gomock.Any()).Return(resp, nil).AnyTimes()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	cfg.RawURL = "https://example.com"
	phase := NewBootstrapPhase(cfg, log, mockHTTP, mockFS)

	progressChan := make(chan ProgressUpdate, 100)
	result := phase.Execute(progressChan)

	if !result.Success {
		t.Errorf("Execute() Success = %v, want true", result.Success)
	}
	if result.Message != "All configuration files downloaded" {
		t.Errorf("Execute() Message = %q, want 'All configuration files downloaded'", result.Message)
	}
}

// TestBootstrapPhaseRollback tests rollback cleanup
func TestBootstrapPhaseRollback(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mocks.MockFileSystem)
		wantErr     bool
		errContains string
	}{
		{
			name: "Rollback - success",
			setupMocks: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().RemoveAll(config.DefaultInstallDir).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Rollback - directory removal failed",
			setupMocks: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().RemoveAll(config.DefaultInstallDir).Return(fmt.Errorf("permission denied")).Times(1)
			},
			wantErr:     true,
			errContains: "failed to remove install directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTP := mocks.NewMockHTTPClient(ctrl)
			mockFS := mocks.NewMockFileSystem(ctrl)

			tt.setupMocks(mockFS)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewBootstrapPhase(cfg, log, mockHTTP, mockFS)

			err = phase.Rollback()
			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if errMsg := err.Error(); tt.errContains != "" && !containsSubstring(errMsg, tt.errContains) {
					t.Errorf("Rollback() error = %q, should contain %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

// MockWriteCloser is a mock implementation of io.WriteCloser
type MockWriteCloser struct {
	buf bytes.Buffer
}

func NewMockWriteCloser() *MockWriteCloser {
	return &MockWriteCloser{}
}

func (m *MockWriteCloser) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func (m *MockWriteCloser) Close() error {
	return nil
}

// containsSubstring helper function (also defined in preflight_test.go)
func containsSubstring(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
