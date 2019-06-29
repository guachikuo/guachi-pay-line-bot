package linebot

import (
	"strconv"

	"github.com/sirupsen/logrus"
)

const (
	// Brown, Cony & Sally
	stickerPackageID1 = "11537"
	// CHOCO & Friends
	stickerPackageID2 = "11538"
	// UNIVERSTAR BT21
	stickerPackageID3 = "11539"
)

var (
	// https://developers.line.biz/media/messaging-api/sticker_list.pdf
	stickerList = map[string]map[string]struct{}{}
)

func init() {
	// https://developers.line.biz/media/messaging-api/sticker_list.pdf
	stickerPackage := map[string]struct{}{}
	for i := 52002734; i <= 52002770; i++ {
		stickerPackage[strconv.FormatInt(int64(i), 10)] = struct{}{}
	}
	for i := 52002777; i <= 52002779; i++ {
		stickerPackage[strconv.FormatInt(int64(i), 10)] = struct{}{}
	}
	stickerList[stickerPackageID1] = stickerPackage

	stickerPackage = map[string]struct{}{}
	for i := 51626494; i <= 51626533; i++ {
		stickerPackage[strconv.FormatInt(int64(i), 10)] = struct{}{}
	}
	stickerList[stickerPackageID2] = stickerPackage

	stickerPackage = map[string]struct{}{}
	for i := 52114110; i <= 52114149; i++ {
		stickerPackage[strconv.FormatInt(int64(i), 10)] = struct{}{}
	}
	stickerList[stickerPackageID3] = stickerPackage
}

func getSticker() (packageID, stickerID string) {
	for packageID, stickers := range stickerList {
		for stickerID := range stickers {
			logrus.WithFields(logrus.Fields{
				"packageID": packageID,
				"stickerID": stickerID,
			}).Info("sticker is selected")
			return packageID, stickerID
		}
	}
	return "", ""
}
