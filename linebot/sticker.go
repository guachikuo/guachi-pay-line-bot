package linebot

const (
	// Brown, Cony & Sally
	stickerPackageID1 = "11537"
	// CHOCO & Friends
	stickerPackageID2 = "11538"
)

var (
	// https://developers.line.biz/media/messaging-api/sticker_list.pdf
	stickerList = map[string]map[string]string{
		stickerPackageID1: map[string]string{
			"52002735": "Cony is happy",
			"52002758": "Cony makes a funny face",
			"52002755": "Cony is begging for forgiveness",
			"52002767": "Brown is angry",
			"52002744": "Brown is confused",
		},
		stickerPackageID2: map[string]string{
			"51626501": "Moon thumbs up",
			"51626504": "Moon is laughing",
			"51626522": "Moon is crying",
			"51626511": "Boss is shocked",
			"51626507": "Moon is blowing the trumpet",
		},
	}
)

func getSticker() (packageID, stickerID string) {
	for packageID, stickers := range stickerList {
		for stickerID := range stickers {
			return packageID, stickerID
		}
	}
	return "", ""
}
