package slice

func DistinctFromINT64Slice(sl *[]int64) {
	slMap := make(map[int64]struct{}, len(*sl))

	ind := 0

	ln := len(*sl)

	for {
		if ind == ln {
			break
		}

		if _, ok := slMap[(*sl)[ind]]; ok {
			*sl = append((*sl)[:ind], (*sl)[ind+1:]...)

			ln--

			continue
		}

		slMap[(*sl)[ind]] = struct{}{}

		ind++
	}
}

func DistinctFromINT32Slice(sl *[]int) {
	slMap := make(map[int]struct{}, len(*sl))

	ind := 0

	ln := len(*sl)

	for {
		if ind == ln {
			break
		}

		if _, ok := slMap[(*sl)[ind]]; ok {
			*sl = append((*sl)[:ind], (*sl)[ind+1:]...)

			ln--

			continue
		}

		slMap[(*sl)[ind]] = struct{}{}

		ind++
	}
}
