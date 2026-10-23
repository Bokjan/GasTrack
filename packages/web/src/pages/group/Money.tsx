import { Tooltip } from 'antd';
import { InfoCircleOutlined } from '@ant-design/icons';
import { formatCurrency } from '@gastrack/shared';

/** 金额展示；外币金额（如加油站单价）附带换算为用户币种的参考值 */
export default function Money({ amount, currency, reference }: { amount: number; currency: string; reference?: string }) {
  return (
    <span>
      {formatCurrency(amount, currency)}
      {reference && (
        <Tooltip title={reference}>
          <InfoCircleOutlined style={{ marginLeft: 4, fontSize: 12, color: 'var(--gt-text-tertiary)', cursor: 'help' }} />
        </Tooltip>
      )}
    </span>
  );
}
