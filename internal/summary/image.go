package summary

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"path"
	"strconv"

	"github.com/gabehf/koito/internal/cfg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/webp"
)

type Summary struct {
	Title          string
	TopArtistImage string
	TopArtists     []struct {
		Name            string
		Plays           int
		MinutesListened int
	}
	TopAlbumImage string
	TopAlbums     []struct {
		Title           string
		Plays           int
		MinutesListened int
	}
	TopTrackImage string
	TopTracks     []struct {
		Title           string
		Plays           int
		MinutesListened int
	}
	MinutesListened int
	Plays           int
	AvgPlaysPerDay  float32
	UniqueTracks    int32
	UniqueAlbums    int32
	UniqueArtists   int32
	NewTracks       int32
	NewAlbums       int32
	NewArtists      int32
}

var (
	assetPath         = path.Join("..", "..", "assets")
	titleFontPath     = path.Join(assetPath, "LeagueSpartan-Medium.ttf")
	textFontPath      = path.Join(assetPath, "Jost-Regular.ttf")
	paddingLg         = 30
	paddingMd         = 20
	paddingSm         = 6
	featuredImageSize = 180
	titleFontSize     = 48.0
	textFontSize      = 16.0
	featureTextStart  = paddingLg + paddingMd + featuredImageSize
)

// lots of code borrowed from https://medium.com/@daniel.ruizcamacho/how-to-create-an-image-in-golang-step-by-step-4416affe088f
func GenerateImage(summary *Summary) error {
	base := image.NewRGBA(image.Rect(0, 0, 750, 1100))
	draw.Draw(base, base.Bounds(), image.NewUniform(color.Black), image.Pt(0, 0), draw.Over)

	file, err := os.Create(path.Join(cfg.ConfigDir(), "summary.png"))
	if err != nil {
		return fmt.Errorf("GenerateImage: %w", err)
	}
	defer file.Close()

	// add title
	if err := addText(base, summary.Title, "", image.Pt(paddingLg, 60), titleFontPath, titleFontSize); err != nil {
		return fmt.Errorf("GenerateImage: %w", err)
	}
	// add images
	if err := addImage(base, summary.TopArtistImage, image.Pt(-paddingLg, -120), featuredImageSize); err != nil {
		return fmt.Errorf("GenerateImage: %w", err)
	}
	if err := addImage(base, summary.TopArtistImage, image.Pt(-paddingLg, -120-(featuredImageSize+paddingLg)), featuredImageSize); err != nil {
		return fmt.Errorf("GenerateImage: %w", err)
	}
	if err := addImage(base, summary.TopArtistImage, image.Pt(-paddingLg, -120-(featuredImageSize+paddingLg)*2), featuredImageSize); err != nil {
		return fmt.Errorf("GenerateImage: %w", err)
	}
	// top artists text
	if err := addText(base, "Top Artists", "", image.Pt(featureTextStart, 132), textFontPath, textFontSize); err != nil {
		return fmt.Errorf("GenerateImage: %w", err)
	}
	for rank, artist := range summary.TopArtists {
		if rank == 0 {
			if err := addText(base, artist.Name, strconv.Itoa(artist.Plays)+" plays", image.Pt(featureTextStart, featuredImageSize+10), titleFontPath, titleFontSize); err != nil {
				return fmt.Errorf("GenerateImage: %w", err)
			}
		} else {
			if err := addText(base, artist.Name, strconv.Itoa(artist.Plays)+" plays", image.Pt(featureTextStart, 210+(rank*(int(textFontSize)+paddingSm))), textFontPath, textFontSize); err != nil {
				return fmt.Errorf("GenerateImage: %w", err)
			}
		}
	}
	// top albums text
	if err := addText(base, "Top Albums", "", image.Pt(featureTextStart, 132+featuredImageSize+paddingLg), textFontPath, textFontSize); err != nil {
		return fmt.Errorf("GenerateImage: %w", err)
	}
	for rank, album := range summary.TopAlbums {
		if rank == 0 {
			if err := addText(base, album.Title, strconv.Itoa(album.Plays)+" plays", image.Pt(featureTextStart, featuredImageSize+10), titleFontPath, titleFontSize); err != nil {
				return fmt.Errorf("GenerateImage: %w", err)
			}
		} else {
			if err := addText(base, album.Title, strconv.Itoa(album.Plays)+" plays", image.Pt(featureTextStart, 210+(rank*(int(textFontSize)+paddingSm))), textFontPath, textFontSize); err != nil {
				return fmt.Errorf("GenerateImage: %w", err)
			}
		}
	}
	// top tracks text

	// stats text

	if err := png.Encode(file, base); err != nil {
		return fmt.Errorf("GenerateImage: png.Encode: %w", err)
	}
	return nil
}

func addImage(baseImage *image.RGBA, path string, point image.Point, height int) error {
	templateFile, err := os.Open(path)
	if err != nil {
		return err
	}

	template, _, err := image.Decode(templateFile)
	if err != nil {
		return err
	}

	resized := resize(template, height, height)

	draw.Draw(baseImage, baseImage.Bounds(), resized, point, draw.Over)

	return nil
}

func addText(baseImage *image.RGBA, text, subtext string, point image.Point, fontFile string, fontSize float64) error {
	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		return err
	}

	ttf, err := opentype.Parse(fontBytes)
	if err != nil {
		return err
	}

	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return err
	}

	drawer := &font.Drawer{
		Dst:  baseImage,
		Src:  image.NewUniform(color.White),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(point.X),
			Y: fixed.I(point.Y),
		},
	}

	drawer.DrawString(text)
	if subtext != "" {
		face, err = opentype.NewFace(ttf, &opentype.FaceOptions{
			Size:    textFontSize,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		drawer.Face = face
		if err != nil {
			return err
		}
		drawer.Src = image.NewUniform(color.RGBA{200, 200, 200, 255})
		drawer.DrawString(" - ")
		drawer.DrawString(subtext)
	}

	return nil
}

func resize(m image.Image, w, h int) *image.RGBA {
	if w < 0 || h < 0 {
		return nil
	}
	r := m.Bounds()
	if w == 0 || h == 0 || r.Dx() <= 0 || r.Dy() <= 0 {
		return image.NewRGBA(image.Rect(0, 0, w, h))
	}
	curw, curh := r.Dx(), r.Dy()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			// Get a source pixel.
			subx := x * curw / w
			suby := y * curh / h
			r32, g32, b32, a32 := m.At(subx, suby).RGBA()
			r := uint8(r32 >> 8)
			g := uint8(g32 >> 8)
			b := uint8(b32 >> 8)
			a := uint8(a32 >> 8)
			img.SetRGBA(x, y, color.RGBA{r, g, b, a})
		}
	}
	return img
}
