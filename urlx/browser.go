package urlx

import "context"

var (
	// MacEdgeAgent Mac Edge 浏览器的 UserAgent
	MacEdgeAgent = UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36 Edg/96.0.1054.43")

	// IPhoneAgent IPhone Edge 浏览器的 UserAgent
	IPhoneAgent = UserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1 Edg/96.0.4664.55")

	// WindowsEdgeAgent Windows Edge 浏览器的 UserAgent
	WindowsEdgeAgent = UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36 Edg/96.0.1054.43")
)

// NewBrowser 浏览器
func NewBrowser(ctx context.Context, userAgent HeaderOption) *Request {
	return New(ctx).ProcessWith(CharsetDecode, DecompressionBody).HeaderWith(AcceptAllEncodings, AcceptHTML, AcceptChinese, NoCache, userAgent)
}

// MacEdge Mac Edge 浏览器
func MacEdge(ctx context.Context) *Request {
	return NewBrowser(ctx, MacEdgeAgent)
}

// WindowsEdge Windows Edge 浏览器
func WindowsEdge(ctx context.Context) *Request {
	return NewBrowser(ctx, WindowsEdgeAgent)
}

// IPhoneEdge IPhone Edge 浏览器
func IPhoneEdge(ctx context.Context) *Request {
	return NewBrowser(ctx, IPhoneAgent)
}
