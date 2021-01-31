package main

import (
	"fmt"
	"os"
	"strconv"
)

func isError(err error) bool { //for handling errors while creating or writing file

	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func createFile(path string) { //creates a file in a path

	//path is the location at which the file will be created

	var _, err = os.Stat(path) //check if path exists

	if os.IsNotExist(err) {
		var file, err = os.Create(path) // create file if not exists
		if isError(err) {

			return
		}
		defer file.Close()
	}

	fmt.Println("File Created Successfully", path)
}

func createPath(i int) string {
	//path generation routine
	path := "./files/"
	fileN := "text"
	No := strconv.Itoa(i)
	fileN = fileN + No + ".txt"
	path = path + fileN
	return path
}

func writeFile(path string) {

	var file, err = os.OpenFile(path, os.O_RDWR, 0644) //opening the file
	if isError(err) {
		return
	}
	defer file.Close()

	_, err = file.WriteString(path) //writing the path of the file into the file
	if isError(err) {
		return
	}

	err = file.Sync() //saving
	if isError(err) {
		return
	}

	fmt.Println("File Updated Successfully.")
}

func main() {

	var path string
	var count int
	fmt.Println("Enter the number of files you want to create: ")
	fmt.Scanln(&count)

	for i := 0; i < count; i++ {
		//creating n number of files
		path = createPath(i) //creating a new path
		createFile(path)     //creating the file
		writeFile(path)      //writing content into the file
	}

}
