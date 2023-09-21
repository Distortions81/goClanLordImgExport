# goClanLordImgExport
This utility is for exporting images from the Clan Lord MMO game.<br>
https://www.deltatao.com/clanlord/<br>
<br>
Easy way: Copy CL_Images into directory and run CLImgExport binary for your platform.<br>
(binaries available under "releases" tab to the right)<br>
<br>
Harder way:<br>
Download golang 1.20 or higher<br>
Copy CL_Images into the directory with the code<br>
Go to directory and run go get, go build, and run CLImgExport.<br>
Alt: Go to dir, go run .<br>
<br>
Note:<br>
The program will create a directory called out, and will save 7000+ PNG images from the CL_Images file.<br>
If found in the images file, the image name will be included as well (id-0123-item name.png).<br>
This software is unlicence (completely free).<br>
<br>
This project is based on: https://github.com/mpolney/clext/<br>
