package game

import (
	"embed"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"regexp"
	"runtime"
	"strconv"
	"sync"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/sirupsen/logrus"
)

//go:embed maps
var MintMaps embed.FS

type MintInfo struct {
	StageId int
	Floor   int
}

func (i MintInfo) MapImage() (*image.RGBA, error) {
	var kind string
	switch i.StageId {
	case CoinMintId:
		kind = "coin"
	case DollarMintId:
		kind = "dollar"
	case BullionMintId:
		kind = "bullion"
	default:
		return nil, fmt.Errorf("unknown stage id: %d", i.StageId)
	}

	path := fmt.Sprintf("maps/%s_%02d.png", kind, i.Floor+1)
	f, err := MintMaps.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening map: %w", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, img.Bounds(), img, image.Point{}, draw.Src)
	return rgba, nil
}

var infoRegex = regexp.MustCompile(`^:.*: stageId (\d+), floor (\d+), \[[\d\s,]+\]$`)

func ScanForMintInfo(logs <-chan string) (MintInfo, error) {
	for line := range logs {
		if matches := infoRegex.FindStringSubmatch(line); matches != nil {
			stageId, _ := strconv.Atoi(matches[1])
			floor, _ := strconv.Atoi(matches[2])
			return MintInfo{StageId: stageId, Floor: floor}, nil
		}
	}
	return MintInfo{}, errors.New("no mint info found")
}

func (m MintInfo) String() string {
	var kind string
	switch m.StageId {
	case CoinMintId:
		kind = "Coin"
	case DollarMintId:
		kind = "Dollar"
	case BullionMintId:
		kind = "Bullion"
	default:
		kind = "Unknown"
	}
	return fmt.Sprintf("%s Mint, Floor %d", kind, m.Floor+1)
}

var (
	glfwInitialized = make(chan struct{})
	glfwTasks       = make(chan func() error)
	glfwDone        = make(chan struct{})
)

var runGLFWOnce sync.Once

var mainGoroutineID uint64

func init() {
	runtime.LockOSThread()
}

func RunGLFW() {
	runGLFWOnce.Do(func() {
		defer close(glfwDone)

		if err := glfw.Init(); err != nil {
			panic(err)
		}
		close(glfwInitialized)
		defer glfw.Terminate()

		for task := range glfwTasks {
			if err := task(); err != nil {
				logrus.Error("error during glfw task:", err)
			}
		}
	})
}

func ShutdownGLFW() {
	<-glfwInitialized
	close(glfwTasks)
	glfw.PostEmptyEvent()
	<-glfwDone
}

func ShowMintInfo(info MintInfo) error {
	img, err := info.MapImage()
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	glfwTasks <- func() error {
		glfw.WindowHint(glfw.Visible, glfw.True)
		glfw.WindowHint(glfw.Resizable, glfw.False)
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 6)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

		window, err := glfw.CreateWindow(bounds.Dx(), bounds.Dy(), info.String(), nil, nil)
		if err != nil {
			return err
		}
		window.MakeContextCurrent()

		gl.Init()
		gl.Enable(gl.TEXTURE_2D)

		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(bounds.Dx()), int32(bounds.Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		var framebuffer uint32
		gl.GenFramebuffers(1, &framebuffer)
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer)
		gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

		for !window.ShouldClose() {
			gl.Clear(gl.COLOR_BUFFER_BIT)

			gl.Viewport(0, 0, int32(bounds.Dx()), int32(bounds.Dy()))

			gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
			// blit the framebuffer upside down since the texture origin is in the bottom left
			gl.BlitFramebuffer(0, int32(bounds.Dy()), int32(bounds.Dx()), 0, 0, 0, int32(bounds.Dx()), int32(bounds.Dy()), gl.COLOR_BUFFER_BIT, gl.LINEAR)

			window.SwapBuffers()
			glfw.WaitEvents()
		}

		window.Destroy()
		return nil
	}

	return nil
}
