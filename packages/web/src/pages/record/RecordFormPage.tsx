import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  InputNumber,
  DatePicker,
  Switch,
  Button,
  Space,
  message,
  Spin,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { fuelRecordApi, useAuthStore } from '@gastrack/shared';
import type { CreateFuelRecordRequest } from '@gastrack/shared';
import dayjs from 'dayjs';

export default function RecordFormPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId, recordId } = useParams<{
    vehicleId: string;
    recordId: string;
  }>();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const user = useAuthStore((s) => s.user);

  const isEdit = !!recordId;

  useEffect(() => {
    if (isEdit && vehicleId && recordId) {
      setFetching(true);
      fuelRecordApi
        .getById(vehicleId, recordId)
        .then(({ data }) => {
          const record = data.data;
          form.setFieldsValue({
            ...record,
            refuel_date: dayjs(record.refuel_date),
          });
        })
        .catch(() => message.error(t('common.error')))
        .finally(() => setFetching(false));
    }
  }, [vehicleId, recordId]);

  // 自动计算总费用
  const handleCalcTotal = () => {
    const amount = form.getFieldValue('fuel_amount');
    const price = form.getFieldValue('unit_price');
    if (amount && price) {
      form.setFieldValue('total_cost', Math.round(amount * price * 100) / 100);
    }
  };

  const onFinish = async (values: CreateFuelRecordRequest & { refuel_date: dayjs.Dayjs }) => {
    setLoading(true);
    const payload: CreateFuelRecordRequest = {
      ...values,
      refuel_date: values.refuel_date.format('YYYY-MM-DDTHH:mm:ssZ'),
      currency_code: user?.currency_code || 'CNY',
    };

    try {
      if (isEdit) {
        await fuelRecordApi.update(vehicleId!, recordId!, payload);
      } else {
        await fuelRecordApi.create(vehicleId!, payload);
      }
      message.success(t('common.success'));
      navigate(`/vehicles/${vehicleId}/records`);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      message.error(error.response?.data?.message || t('common.error'));
    } finally {
      setLoading(false);
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
            onClick={() => navigate(`/vehicles/${vehicleId}/records`)}
          />
          <h2>
            {isEdit ? t('fuelRecord.editRecord') : t('fuelRecord.addRecord')}
          </h2>
        </Space>
      </div>

      <Card style={{ maxWidth: 600 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            refuel_date: dayjs(),
            is_full_tank: true,
          }}
        >
          <Form.Item
            name="refuel_date"
            label={t('fuelRecord.fuelDate')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <DatePicker
              showTime
              style={{ width: '100%' }}
              format="YYYY-MM-DD HH:mm"
            />
          </Form.Item>

          <Form.Item name="station_name" label={t('fuelRecord.station')}>
            <Input placeholder={t('fuelRecord.station')} />
          </Form.Item>

          <Space style={{ width: '100%' }} size="middle">
            <Form.Item
              name="fuel_amount"
              label={t('fuelRecord.fuelAmount')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <InputNumber
                min={0.01}
                step={0.01}
                style={{ width: '100%' }}
                suffix="L"
                onChange={handleCalcTotal}
              />
            </Form.Item>

            <Form.Item
              name="unit_price"
              label={t('fuelRecord.pricePerUnit')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <InputNumber
                min={0.01}
                step={0.01}
                style={{ width: '100%' }}
                onChange={handleCalcTotal}
              />
            </Form.Item>
          </Space>

          <Space style={{ width: '100%' }} size="middle">
            <Form.Item
              name="total_cost"
              label={t('fuelRecord.totalCost')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="odometer"
              label={t('fuelRecord.odometer')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <InputNumber
                min={0}
                step={1}
                style={{ width: '100%' }}
                suffix="km"
              />
            </Form.Item>
          </Space>

          <Form.Item
            name="is_full_tank"
            label={t('fuelRecord.isFullTank')}
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item name="note" label={t('fuelRecord.notes')}>
            <Input.TextArea rows={3} placeholder={t('fuelRecord.notes')} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {t('common.save')}
              </Button>
              <Button
                onClick={() => navigate(`/vehicles/${vehicleId}/records`)}
              >
                {t('common.cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
