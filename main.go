package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "os/exec"
    "strconv"
    "strings"

    "github.com/kkdai/youtube/v2"
)

func downloadStream(client *youtube.Client, video *youtube.Video, format *youtube.Format, filename string) error {
    stream, _, err := client.GetStream(video, format)
    if err != nil {
        return err
    }
    defer stream.Close()

    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.Copy(file, stream)
    if err != nil {
        return err
    }

    return nil
}

func findFormatByItag(formats youtube.FormatList, itagNo int) *youtube.Format {
    for _, format := range formats {
        if format.ItagNo == itagNo {
            return &format
        }
    }
    return nil
}

func main() {
    client := &youtube.Client{}
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter the YouTube video URL: ")
    url, _ := reader.ReadString('\n')
    url = strings.TrimSpace(url) 

    video, err := client.GetVideo(url)
    if err != nil {
        panic(fmt.Sprintf("Failed to get video info: %v", err))
    }

    fmt.Println("Available formats:")
    for _, f := range video.Formats {
        fmt.Printf("Quality: %s, MimeType: %s, Itag: %d, AudioChannels: %d\n", f.Quality, f.MimeType, f.ItagNo, f.AudioChannels)
    }

    var videoItagStr, audioItagStr string
    fmt.Print("Enter the desired video Itag: ")
    fmt.Scanln(&videoItagStr)
    fmt.Print("Enter the desired audio Itag: ")
    fmt.Scanln(&audioItagStr)

    videoItag, err := strconv.Atoi(videoItagStr)
    if err != nil {
        panic("Invalid video Itag")
    }
    audioItag, err := strconv.Atoi(audioItagStr)
    if err != nil {
        panic("Invalid audio Itag")
    }

    videoFormat := findFormatByItag(video.Formats, videoItag)
    if videoFormat == nil {
        panic(fmt.Sprintf("Could not find video format with itag %d", videoItag))
    }

    audioFormat := findFormatByItag(video.Formats, audioItag)
    if audioFormat == nil {
        panic(fmt.Sprintf("Could not find audio format with itag %d", audioItag))
    }

    fmt.Println("Downloading video...")
    err = downloadStream(client, video, videoFormat, "video.mp4")
    if err != nil {
        panic(fmt.Sprintf("Failed to download video: %v", err))
    }

    fmt.Println("Downloading audio...")
    err = downloadStream(client, video, audioFormat, "audio.m4a")
    if err != nil {
        panic(fmt.Sprintf("Failed to download audio: %v", err))
    }

    fmt.Println("Combining video and audio...")
    cmd := exec.Command("ffmpeg", "-i", "video.mp4", "-i", "audio.m4a", "-c:v", "copy", "-c:a", "aac", "output.mp4")
    err = cmd.Run()
    if err != nil {
        panic(fmt.Sprintf("Failed to combine video and audio: %v", err))
    }

    fmt.Println("Video downloaded and combined successfully as output.mp4")
}
