# 概览
自动打包，并可选择是否上传到蒲公英。

由于目前大多数商业项目使用 Cocoapods 来管理第三方库，所以目前该自动打包程序仅支持 .xcworkspace 工程。

# 使用前设置需要修改
Pgyer_APIKey： 蒲公英 API key

workspaceName: 打包的工程，如`demo.xcworkspace`。

schemeName: 打包的 scheme 名字。

archivePath: 打包路径。如果设置为`~/Desktop/demo`，那么最终打包完的文件路径为`~/Desktop/demo.xcarchive`。

ipaPath：导出的 ipa 所在的文件夹。如果设置为`~/Desktop/demo-ipa`，导出后，ipa 文件路径为`~/Desktop/demo-ipa/xxx.ipa`。

exportOptionsPath：导出选项配置文件路径。关于配置文件各个选项的配置，可通过`xcodebuild -help`来了解。

# 使用说明
可以通过`go run archive.go`，或者 build 后通过`./archive`来运行程序。

可通过`go run archive.go -help`，或通过`./archive -help`来了解该程序的更多选项。

实例一，在`Debug`模式下打包，上传到蒲公英的更新说明为空字符串：
```
./archive
```

实例二，在`Release`模式下打包，上传到蒲公英的更新说明为`添加了一些惊为天人的新特性`：
```
./archive -configuration "Release" -uploadDescription "添加了一些惊为天人的新特性"
```