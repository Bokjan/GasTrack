import { Alert } from 'antd';
import { useTranslation } from 'react-i18next';

interface UnconvertedNoticeProps {
  /** 后端返回的无参考汇率、未计入换算总额的币种 */
  currencies?: (string[] | undefined)[];
  style?: React.CSSProperties;
}

/** 提示部分币种因缺少汇率未计入总额 */
export default function UnconvertedNotice({ currencies, style }: UnconvertedNoticeProps) {
  const { t } = useTranslation();
  const list = [...new Set((currencies || []).flatMap((c) => c || []))];
  if (list.length === 0) return null;
  return (
    <Alert
      type="warning"
      showIcon
      style={{ marginBottom: 16, ...style }}
      message={t('exchangeRate.unconvertedNotice', { currencies: list.join(', ') })}
    />
  );
}
