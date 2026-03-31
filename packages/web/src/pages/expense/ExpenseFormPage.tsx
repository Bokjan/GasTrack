import { useEffect, useState, useMemo } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  InputNumber,
  DatePicker,
  Select,
  Button,
  Space,
  message,
  Spin,
  AutoComplete,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  expenseApi,
  reminderApi,
  useAuthStore,
  CURRENCIES,
} from '@gastrack/shared';
import type { CreateExpenseRequest, Reminder, ExpenseCategory, MaintenanceCategory } from '@gastrack/shared';
import dayjs from 'dayjs';

const EXPENSE_CATEGORIES: ExpenseCategory[] = [
  'maintenance', 'repair', 'insurance', 'parking', 'toll',
  'car_wash', 'inspection', 'parts', 'fine', 'other',
];

const MAINTENANCE_CATEGORIES: MaintenanceCategory[] = [
  'oil_change', 'tire_rotation', 'brake_pads', 'air_filter',
  'transmission', 'coolant', 'spark_plugs', 'battery',
  'tire_replace', 'inspection', 'custom',
];

export default function ExpenseFormPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId, expenseId } = useParams<{
    vehicleId: string;
    expenseId: string;
  }>();
  const [searchParams] = useSearchParams();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [vendorSuggestions, setVendorSuggestions] = useState<string[]>([]);
  const [vendorSearch, setVendorSearch] = useState('');
  const [reminders, setReminders] = useState<Reminder[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const user = useAuthStore((s) => s.user);

  const isEdit = !!expenseId;
  const isImperial = user?.unit_system === 'imperial';
  const defaultDistanceUnit = isImperial ? 'mi' : 'km';
  const defaultCurrency = user?.currency_code || 'CNY';

  // 从 URL 查询参数预填（从提醒页"完成并记账"跳转时）
  const preCategory = searchParams.get('category') || 'maintenance';
  const preMaintenanceCategory = searchParams.get('maintenance_category') || '';
  const preTitle = searchParams.get('title') || '';
  const preReminderId = searchParams.get('reminder_id') || '';

  // 初始化 selectedCategory（来自 URL 参数或默认值）
  useEffect(() => {
    if (!isEdit && preCategory) {
      setSelectedCategory(preCategory);
    }
  }, []);

  // 商家建议筛选
  const vendorOptions = useMemo(() => {
    if (!vendorSearch) {
      return vendorSuggestions.map((name) => ({ value: name, label: name }));
    }
    const lower = vendorSearch.toLowerCase();
    return vendorSuggestions
      .filter((name) => name.toLowerCase().includes(lower))
      .map((name) => ({ value: name, label: name }));
  }, [vendorSuggestions, vendorSearch]);

  useEffect(() => {
    if (vehicleId) {
      // 加载商家建议
      expenseApi.getVendorSuggestions(vehicleId).then(({ data }) => {
        setVendorSuggestions(data.data || []);
      }).catch(() => { /* ignore */ });

      // 加载保养提醒列表（用于联动选择）
      reminderApi.list().then(({ data }) => {
        // 只显示当前车辆的提醒
        const vehicleReminders = (data.data || []).filter(
          (r: Reminder) => r.vehicle_id === vehicleId && r.is_enabled
        );
        setReminders(vehicleReminders);
      }).catch(() => { /* ignore */ });
    }
  }, [vehicleId]);

  useEffect(() => {
    if (isEdit && vehicleId && expenseId) {
      setFetching(true);
      expenseApi
        .getById(vehicleId, expenseId)
        .then(({ data }) => {
          const record = data.data;
          setSelectedCategory(record.category);
          form.setFieldsValue({
            ...record,
            expense_date: dayjs(record.expense_date),
          });
        })
        .catch(() => message.error(t('common.error')))
        .finally(() => setFetching(false));
    }
  }, [vehicleId, expenseId]);

  const onFinish = async (values: CreateExpenseRequest & { expense_date: dayjs.Dayjs }) => {
    if (!vehicleId) return;

    setLoading(true);
    const payload: CreateExpenseRequest = {
      ...values,
      expense_date: values.expense_date.format('YYYY-MM-DDTHH:mm:ssZ'),
      currency_code: values.currency_code || defaultCurrency,
      distance_unit: values.distance_unit || defaultDistanceUnit,
    };

    // 非保养类别清除 maintenance_category
    if (payload.category !== 'maintenance') {
      payload.maintenance_category = undefined;
    }

    try {
      if (isEdit) {
        await expenseApi.update(vehicleId, expenseId!, payload);
        message.success(t('expense.updateSuccess'));
      } else {
        await expenseApi.create(vehicleId, payload);
        message.success(t('expense.createSuccess'));
      }
      navigate(`/vehicles/${vehicleId}/expenses`);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      message.error(error.response?.data?.message || t('common.error'));
    } finally {
      setLoading(false);
    }
  };

  const handleCategoryChange = (value: string) => {
    setSelectedCategory(value);
    if (value !== 'maintenance') {
      form.setFieldValue('maintenance_category', undefined);
      form.setFieldValue('reminder_id', undefined);
    }
  };

  if (fetching) {
    return (
      <div style={{ textAlign: 'center', padding: 80 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/vehicles/${vehicleId}/expenses`)}
          />
          <h2>{isEdit ? t('expense.edit') : t('expense.add')}</h2>
        </Space>
      </div>

      <Card style={{ maxWidth: 600 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            expense_date: dayjs(),
            currency_code: defaultCurrency,
            distance_unit: defaultDistanceUnit,
            category: preCategory,
            maintenance_category: preMaintenanceCategory || undefined,
            title: preTitle,
            reminder_id: preReminderId || undefined,
          }}
        >
          {/* 类别 */}
          <Form.Item
            name="category"
            label={t('expense.category.label')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Select onChange={handleCategoryChange}>
              {EXPENSE_CATEGORIES.map((cat) => (
                <Select.Option key={cat} value={cat}>
                  {t(`expense.category.${cat}`)}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          {/* 保养子类别（仅 category=maintenance 时显示） */}
          {selectedCategory === 'maintenance' && (
            <Form.Item
              name="maintenance_category"
              label={t('expense.maintenanceCategory.label')}
              rules={[{ required: true, message: t('common.required') }]}
            >
              <Select>
                {MAINTENANCE_CATEGORIES.map((cat) => (
                  <Select.Option key={cat} value={cat}>
                    {t(`expense.maintenanceCategory.${cat}`)}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
          )}

          {/* 标题 */}
          <Form.Item
            name="title"
            label={t('expense.field.title')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input maxLength={200} />
          </Form.Item>

          {/* 金额 */}
          <Form.Item
            name="amount"
            label={t('expense.field.amount')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <InputNumber
              min={0.01}
              step={0.01}
              style={{ width: '100%' }}
              prefix={CURRENCIES.find((c) => c.value === defaultCurrency)?.symbol || defaultCurrency}
            />
          </Form.Item>
          <Form.Item name="currency_code" hidden><Input /></Form.Item>

          {/* 日期 */}
          <Form.Item
            name="expense_date"
            label={t('expense.field.date')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <DatePicker style={{ width: '100%' }} format="YYYY-MM-DD" />
          </Form.Item>

          {/* 商家/服务商 */}
          <Form.Item name="vendor_name" label={t('expense.field.vendor')}>
            <AutoComplete
              options={vendorOptions}
              onSearch={setVendorSearch}
              placeholder={t('expense.field.vendor')}
              allowClear
              filterOption={false}
            />
          </Form.Item>

          {/* 里程 */}
          <Form.Item name="odometer" label={t('expense.field.odometer')}>
            <InputNumber min={0} step={1} style={{ width: '100%' }} suffix={defaultDistanceUnit} />
          </Form.Item>
          <Form.Item name="distance_unit" hidden><Input /></Form.Item>

          {/* 保养提醒联动（仅 category=maintenance 时） */}
          {selectedCategory === 'maintenance' && reminders.length > 0 && (
            <Form.Item name="reminder_id" label={t('expense.reminderLink.label')}>
              <Select allowClear placeholder={t('expense.reminderLink.none')}>
                {reminders.map((reminder) => (
                  <Select.Option key={reminder.id} value={reminder.id}>
                    {reminder.title} ({t(`expense.maintenanceCategory.${reminder.category}`)})
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
          )}

          {/* 备注 */}
          <Form.Item name="note" label={t('expense.field.note')}>
            <Input.TextArea rows={3} maxLength={1000} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {t('common.save')}
              </Button>
              <Button onClick={() => navigate(`/vehicles/${vehicleId}/expenses`)}>
                {t('common.cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
