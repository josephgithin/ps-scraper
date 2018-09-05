package scrapper

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/kennygrant/sanitize"
)

//PsScrap is a  scrapper
func PsScrap(userName string, password string, courseName string, firstModule string) {

	var err error

	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctxt) // chromedp.WithLog(log.Printf)
	if err != nil {
		log.Fatal(err)
	}

	// run task list
	err = c.Run(ctxt, doLogin(userName, password))
	if err != nil {
		log.Fatal(err)
	}

	urlIndex, urlMap := getClipPlayerUrls(ctxt, c, courseName, firstModule)

	// c.Run(ctxt, chromedp.Click(`#play-control`))
	stateFileName := courseName + ".state"

	lastIndexSaved := -1
	if _, err := os.Stat(stateFileName); os.IsNotExist(err) {
		log.Println("state file dont exist")
		f, err := os.Create(stateFileName)
		if err != nil {
			defer f.Close()
		}

	} else {

		stateFile, err := os.Open(stateFileName)
		if err == nil {
			scanner := bufio.NewScanner(stateFile)
			scanner.Scan()
			data := scanner.Text()
			lastIndexSaved, err = strconv.Atoi(data)
			stateFile.Close()
		} else {
			log.Printf("Unable to read state file")
		}
	}

	lastIndexSaved = lastIndexSaved + 1

	log.Printf(fmt.Sprintf("lastIndex=%v", lastIndexSaved))
	log.Printf(fmt.Sprintf("lenurlIndex=%v", len(urlIndex)))
	for i := lastIndexSaved; i < len(urlIndex); i++ {

		urlKey := urlIndex[i]
		url := urlMap[urlKey]
		downloadClip(ctxt, c, url, courseName, urlKey)
		err := os.Truncate(stateFileName, 0)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(stateFileName, []byte(strconv.Itoa(i)), 0666)

	}

	var title string
	c.Run(ctxt, chromedp.InnerHTML("#course-title-link", &title, chromedp.NodeVisible, chromedp.ByID))
	fmt.Println(title)

	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func doLogin(userName string, password string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(`https://app.pluralsight.com/id?redirectTo=%2Fid%2Fdashboard`),
		chromedp.WaitVisible(`#Username`),
		chromedp.SendKeys(`#Username`, userName),
		chromedp.WaitVisible(`#Password`),
		chromedp.Sleep(2 * time.Second),
		chromedp.SendKeys(`#Password`, "whitedev"),
		chromedp.Sleep(2 * time.Second),
		chromedp.Click(`#login`),
		chromedp.WaitVisible(`#prism-search-input`),
	}
}

func getClipPlayerUrls(ctxt context.Context, c *chromedp.CDP, courseName string, firstModule string) (map[int]string, map[string]string) {

	baseurl := "https://app.pluralsight.com/player"
	courseURL := fmt.Sprintf(`%v?course=%v`, baseurl, courseName)

	waitToLoad(ctxt, c, courseURL)
	c.Run(ctxt, chromedp.Sleep(2*time.Second))
	c.Run(ctxt, chromedp.Click(`#vjs_video_3`, chromedp.NodeVisible, chromedp.ByID))

	var moduleCount int

	c.Run(ctxt, chromedp.Sleep(1*time.Second))
	c.Run(ctxt, chromedp.EvaluateAsDevTools("document.getElementsByClassName('module').length", &moduleCount))
	log.Printf("moduleCount=" + string(moduleCount))
	c.Run(ctxt, chromedp.Sleep(1*time.Second))
	moduleClipMap := make(map[string]int)

	urlIndex := make(map[int]string)
	urlmap := make(map[string]string)

	re := regexp.MustCompile("[0-9]+")
	counter := 0
	firstModuleIndex, err := strconv.Atoi(re.FindAllString(firstModule, -1)[0])

	if err != nil {
		log.Println("Error parsing firstmoduleIndex")
	} else {

		for i := 0; i < moduleCount; i++ {

			log.Println(i)

			var isOpen bool
			c.Run(ctxt, chromedp.Sleep(1*time.Second))

			c.Run(ctxt, chromedp.EvaluateAsDevTools(fmt.Sprintf("document.getElementsByClassName('module')[%v].classList.contains('open')", i), &isOpen))

			if isOpen != true {
				log.Println("isNotOpen")
				c.Run(ctxt, chromedp.EvaluateAsDevTools(fmt.Sprintf("document.getElementsByClassName('module')[%v].children[0].click()", i), &isOpen))

			}

			c.Run(ctxt, chromedp.Sleep(1*time.Second))

			var clipCount int
			c.Run(ctxt, chromedp.EvaluateAsDevTools(fmt.Sprintf("document.getElementsByClassName('module')[%v].children[1].children.length", i),
				&clipCount))
			moduleClipMap[fmt.Sprintf("m%v", i+firstModuleIndex)] = clipCount
			for j := 0; j < clipCount; j++ {
				clipurl := fmt.Sprintf("%v?course=%v&name=%v&clip=%v", baseurl, courseName,
					courseName+fmt.Sprintf("-m%v", i+firstModuleIndex), j)
				log.Println(clipurl)

				urlIndex[counter] = fmt.Sprintf("%v-%v", i+firstModuleIndex, j)
				urlmap[fmt.Sprintf("%v-%v", i+firstModuleIndex, j)] = clipurl
				counter = counter + 1
			}
		}

	}

	return urlIndex, urlmap
}

func waitToLoad(ctxt context.Context, c *chromedp.CDP, url string) {
	c.Run(ctxt, chromedp.Navigate(url))

	c.Run(ctxt, chromedp.WaitReady(`#play-control`, chromedp.ByID))

	c.Run(ctxt, chromedp.WaitReady(`#module-clip-title`, chromedp.ByID))
	c.Run(ctxt, chromedp.WaitReady(`#vjs_video_3_html5_api`, chromedp.ByID))
	c.Run(ctxt, chromedp.WaitReady(`#tab-table-of-contents`, chromedp.ByID))
	c.Run(ctxt, chromedp.WaitReady(`#vjs_video_3`, chromedp.ByID))
	c.Run(ctxt, chromedp.Sleep(2*time.Second))

}

func downloadClip(ctxt context.Context, c *chromedp.CDP, url string, courseName string, urlkey string) {

	waitToLoad(ctxt, c, url)

	var heading string
	c.Run(ctxt, chromedp.InnerHTML(`#module-clip-title`, &heading, chromedp.NodeVisible, chromedp.ByID))

	fmt.Println(heading)
	var link string
	var ok bool
	c.Run(ctxt, chromedp.AttributeValue(`#vjs_video_3_html5_api`, "src", &link, &ok))

	folderAndPath := strings.Split(heading, ":")

	file := urlkey + "_" + strings.Replace(strings.TrimSpace(folderAndPath[len(folderAndPath)-1]), " ", "_", -1)
	folder := strings.Split(urlkey, "-")[0] + "_" + strings.Replace(strings.TrimSpace(strings.Join(folderAndPath[:len(folderAndPath)-1], " ")), " ", "_", -1)

	fmt.Println(link)
	fileName := fmt.Sprintf("%v%v", file, ".mp4")
	DownloadFile(courseName+"/"+folder, strings.TrimSpace(fileName), link)
}

//DownloadFile Downloads a file
func DownloadFile(folderName string, filepath string, url string) error {
	fmt.Println("Downloading file")
	// Create the file
	folder := sanitize.Path(folderName)
	file := sanitize.Name(filepath)

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, os.ModePerm)
	}

	out, err := os.Create(fmt.Sprintf("%v/%v", folder, file))
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
