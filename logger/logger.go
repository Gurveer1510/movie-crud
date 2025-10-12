package logger
 
import (
    "errors"
    "fmt"
    "os"
    "runtime"
    "time"
)
 
const (
    colorRed   = "\033[31m"
    colorGreen = "\033[32m"
    colorBlue  = "\033[34m"
    colorReset = "\033[0m"
)
 
var logChan chan string
var stopChan chan struct{}
 
const STDIO = 1
const FILE = 2
 
var logType int
var fd *os.File
var path string
var fName string
 
func InitialiseLogger(outputType int, dirPath ...string) error {
    // Sanity checks
    if outputType != STDIO && outputType != FILE {
        return errors.New("Invalid output type")
    }
    if outputType == FILE && len(dirPath) == 0 {
        return errors.New("Directory path not specified")
    }
 
    logChan = make(chan string, 1000)
    stopChan = make(chan struct{}, 1)
    logType = outputType
 
    if outputType == FILE {
        path = dirPath[0]
        fmt.Println("Path provided ", path)
        var err error
        name := time.Now().Format("02-01-2006") + ".log"
        fd, err = os.OpenFile(
            path+"/"+name,
            os.O_APPEND|os.O_CREATE|os.O_WRONLY,
            0666,
        )
        if err != nil {
            return err
        }
        fmt.Println("Writing logs on file ", path+"/"+name)
        fName = name
 
    }
 
    return nil
}
 
func LogEvents() {
    for {
        select {
        case log := <-logChan:
            if logType == STDIO {
                fmt.Println(log)
                continue
            }
 
            // Log day has changed
            if fName != (time.Now().Format("02-01-2006") + ".log") {
                fd.Close()
                name := time.Now().Format("02-01-2006") + ".log"
                var err error
                fd, err = os.OpenFile(
                    path+"/"+name,
                    os.O_APPEND|os.O_CREATE|os.O_WRONLY,
                    0666,
                )
 
                if err != nil {
                    continue
                }
                fName = name
            }
 
            fd.WriteString(log + "\n")
        case <-stopChan:
            close(logChan)
            close(stopChan)
            return
        }
    }
 
}
 
func CleanUp() {
    for len(logChan) != 0 {
 
    }
    stopChan <- struct{}{}
    if logType == FILE {
        fd.Close()
    }
}
 
func Error(v ...any) {
    _, file, line, ok := runtime.Caller(1)
    if ok {
        logChan <- fmt.Sprint(
            colorRed,
            "[ERROR]",
            "[",
            time.Now().Format("15:04:05"),
            "]",
            v,
            " ",
            "Line No:",
            line,
            " ",
            "File:",
            file,
            colorReset,
        )
    }
 
}
 
func Info(v ...any) {
    logChan <- fmt.Sprint(
        colorBlue,
        "[INFO]",
        "[",
        time.Now().Format("15:04:05"),
        "]",
        v,
        colorReset,
    )
}
 
 
 