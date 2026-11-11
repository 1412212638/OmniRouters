package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateWeChatOfficialMenu(t *testing.T) {
	require.NoError(t, ValidateWeChatOfficialMenu(`{"button":[{"name":"官网","type":"view","url":"https://example.com"},{"name":"帮助","type":"click","key":"help","reply":"说明"},{"name":"更多","sub_button":[{"name":"联系","type":"click","key":"contact","reply":"联系我们"}]}]}`))
	require.Error(t, ValidateWeChatOfficialMenu(`{"button":[{"name":"1"},{"name":"2"},{"name":"3"},{"name":"4"}]}`))
	require.Error(t, ValidateWeChatOfficialMenu(`{"button":[{"name":"更多","sub_button":[{"name":"一","sub_button":[{"name":"二","type":"click","key":"deep"}]}]}]}`))
	require.Error(t, ValidateWeChatOfficialMenu(`{"button":[{"name":"官网","type":"view","url":"javascript:alert(1)"}]}`))
	require.Error(t, ValidateWeChatOfficialMenu(`{"button":[{"name":"帮助","type":"click"}]}`))
}
