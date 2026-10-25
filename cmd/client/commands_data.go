package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/aws/smithy-go/ptr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type arbitraryDataResultMsg struct {
	err           error
	arbitraryData []arbitraryData
}

type localArbitraryDataResultMsg struct {
	err                error
	localArbitraryData []localUploadedData
}

type createArbitraryDataResultMsg struct {
	err error
}

type deleteArbitraryDataResultMsg struct {
	err error
}

type reloadArbitraryDataResultMsg struct {
	err error
}
type downloadArbitraryDataResultMsg struct {
	err error
}

type uploadProgressMsg struct {
	sent  int64
	total int64
}

func createArbitraryDataCmd(
	client pb.GoKeeperServiceClient, path string, masterKey string, ch chan<- uploadProgressMsg) tea.Cmd {
	return func() tea.Msg {
		filename := filepath.Base(path)                             // "photo.jpg"
		filetype := strings.TrimPrefix(filepath.Ext(filename), ".") // "jpg"
		// Имя без расширения:
		nameOnly := strings.TrimSuffix(filename, filepath.Ext(filename)) // "photo"

		file, err := os.Open(path)
		if err != nil {
			return createArbitraryDataResultMsg{err: fmt.Errorf("failed to open: %w", err)}
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return createArbitraryDataResultMsg{err: err}
		}

		stream, err := client.CreateArbitraryData(context.Background())
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() == codes.Unauthenticated {
				return refreshTokenErrorMsg{err: err}
			}
			return createArbitraryDataResultMsg{err: err}
		}

		// Если Send вернул io.EOF, настоящую ошибку отдаёт CloseAndRecv.
		sendErr := func(err error) tea.Msg {
			if errors.Is(err, io.EOF) {
				_, err = stream.CloseAndRecv()
			}
			return createArbitraryDataResultMsg{err: err}
		}

		encryptName, err := encrypt([]byte(nameOnly), []byte(masterKey))
		if err != nil {
			return createArbitraryDataResultMsg{err: err}
		}
		encryptOsPath, err := encrypt([]byte(path), []byte(masterKey))
		if err != nil {
			return createArbitraryDataResultMsg{err: err}
		}

		err = stream.Send(pb.CreateArbitraryDataRequest_builder{
			Metadata: pb.FileMetadata_builder{
				Filename: encryptName,
				Type:     filetype,
				Size:     info.Size(),
				OsPath:   encryptOsPath,
			}.Build(),
		}.Build())
		if err != nil {
			return sendErr(err)
		}

		buf := make([]byte, 1024*1024)
		var sent int64
		for {
			n, readErr := file.Read(buf)
			if n > 0 {
				if err := stream.Send(pb.CreateArbitraryDataRequest_builder{
					Chunk: buf[:n],
				}.Build()); err != nil {
					return sendErr(err)
				}
				sent += int64(n)
				ch <- uploadProgressMsg{sent: sent, total: info.Size()}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return createArbitraryDataResultMsg{err: readErr}
			}
		}

		if _, err := stream.CloseAndRecv(); err != nil {
			return createArbitraryDataResultMsg{err: err}
		}
		return createArbitraryDataResultMsg{}
	}
}

func getArbitraryDataCmd(client pb.GoKeeperServiceClient, masterKey string) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetArbitraryData(context.Background(), &pb.ListArbitraryDataRequest{})
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() == codes.Unauthenticated {
				return refreshTokenErrorMsg{err: err}
			}
			return cardsResultMsg{err: err}
		}

		decodingList := make([]arbitraryData, 0, len(resp.GetArbitraryData()))

		for _, v := range resp.GetArbitraryData() {
			var (
				decFilename []byte
				decOsPath   []byte
				err         error
			)
			if len(v.GetName()) > 0 || len(v.GetOsPath()) > 0 {
				decFilename, err = decrypt(v.GetName(), []byte(masterKey))
				if err != nil {
					return cardsResultMsg{err: err}
				}
				decOsPath, err = decrypt(v.GetOsPath(), []byte(masterKey))
				if err != nil {
					return cardsResultMsg{err: err}
				}
			}

			newStruct := arbitraryData{
				id:         v.GetId(),
				filename:   string(decFilename),
				filetype:   v.GetType(),
				status:     v.GetStatus(),
				objectKey:  v.GetObjectKey(),
				osPath:     string(decOsPath),
				clientSize: v.GetClientSize(),
				totalSize:  v.GetTotalSize(),
				createdAt:  v.GetCreatedAt().AsTime(),
			}

			decodingList = append(decodingList, newStruct)
		}
		slog.Error("decodingList", "", decodingList)
		return arbitraryDataResultMsg{err: err, arbitraryData: decodingList}
	}
}

func deleteArbitraryDataCmd(client pb.GoKeeperServiceClient, id int64) tea.Cmd {
	return func() tea.Msg {
		_, err := client.DeleteArbitraryData(context.Background(), pb.DeleteArbitraryDataRequest_builder{
			Id: id,
		}.Build())
		st, _ := status.FromError(err)
		if st.Code() == codes.Unauthenticated {
			return refreshTokenErrorMsg{err: err}
		}
		return deleteArbitraryDataResultMsg{err: err}
	}
}

func reloadArbitraryDataCmd(client pb.GoKeeperServiceClient, id int64, filepath string, ch chan<- uploadProgressMsg) tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(filepath)
		if err != nil {
			return reloadArbitraryDataResultMsg{err: fmt.Errorf("failed to open: %w", err)}
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return reloadArbitraryDataResultMsg{err: err}
		}

		stream, err := client.ReloadArbitraryData(context.Background())
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() == codes.Unauthenticated {
				return refreshTokenErrorMsg{err: err}
			}
			return reloadArbitraryDataResultMsg{err: err}
		}

		// Если Send вернул io.EOF, настоящую ошибку отдаёт CloseAndRecv.
		sendErr := func(err error) tea.Msg {
			if errors.Is(err, io.EOF) {
				_, err = stream.CloseAndRecv()
			}
			return reloadArbitraryDataResultMsg{err: err}
		}

		if err := stream.Send(pb.ReloadArbitraryDataRequest_builder{Id: ptr.Int64(id)}.Build()); err != nil {
			return reloadArbitraryDataResultMsg{err: err}
		}

		buf := make([]byte, 1024*1024)
		var sent int64

		for {
			n, readErr := file.Read(buf)
			if n > 0 {
				if err := stream.Send(pb.ReloadArbitraryDataRequest_builder{
					Chunk: buf[:n],
				}.Build()); err != nil {
					return sendErr(err)
				}

				sent += int64(n)
				ch <- uploadProgressMsg{sent: sent, total: info.Size()}
			}
			if readErr == io.EOF {
				break
			}

			if readErr != nil {
				return reloadArbitraryDataResultMsg{err: readErr}
			}
		}
		if _, err := stream.CloseAndRecv(); err != nil {
			return reloadArbitraryDataResultMsg{err: err}
		}
		return reloadArbitraryDataResultMsg{err: nil}
	}
}

func downloadArbitraryDataCmd(
	client pb.GoKeeperServiceClient, id int64, filename, filetype, destDir string, ch chan<- uploadProgressMsg) tea.Cmd {
	return func() tea.Msg {
		stream, err := client.DownloadArbitraryData(context.Background(), pb.DownloadArbitraryDataRequest_builder{
			Id: id,
		}.Build())

		req, err := stream.Recv()
		if err != nil {
			return downloadArbitraryDataResultMsg{err: err}
		}

		var totalSize int64
		if req.GetMetadata() != nil {
			totalSize = req.GetMetadata().GetSize()
		}

		// создаём директорию, если её нет (если есть, ничего не делает)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return downloadArbitraryDataResultMsg{err: err}
		}
		fullPath := filepath.Join(destDir, filepath.Base(filename+"."+filetype))
		file, err := os.Create(fullPath)
		if err != nil {
			return downloadArbitraryDataResultMsg{err: err}
		}
		defer file.Close()
		var sent int64

		for {
			resp, err := stream.Recv()
			st, _ := status.FromError(err)
			if st.Code() == codes.Unauthenticated {
				return refreshTokenErrorMsg{err: err}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return downloadArbitraryDataResultMsg{err: err}
			}
			if _, err := file.Write(resp.GetChunk()); err != nil {
				return downloadArbitraryDataResultMsg{err: err}
			}

			sent += int64(len(resp.GetChunk()))
			ch <- uploadProgressMsg{sent: sent, total: totalSize}
		}

		return downloadArbitraryDataResultMsg{err: err}
	}
}

func getLocalArbitraryDataCmd() tea.Cmd {
	return func() tea.Msg {
		list, err := scanDir(DOWNLOAD_PATHNAME)
		if err != nil {
			return localArbitraryDataResultMsg{err: err}
		}
		return localArbitraryDataResultMsg{localArbitraryData: list}
	}
}
