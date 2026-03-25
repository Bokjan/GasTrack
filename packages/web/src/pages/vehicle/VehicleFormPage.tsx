import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
  Button,
  Space,
  message,
  Spin,
} from 'antd';
import { useTranslation } from 'react-i18next';
import {
  vehicleApi,
  useVehicleStore,
  FUEL_TYPES,
  VEHICLE_TYPES,
} from '@gastrack/shared';
import type { CreateVehicleRequest, VehicleType } from '@gastrack/shared';

export default function VehicleFormPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [vehicleType, setVehicleType] = useState<VehicleType>('car');
  const { addVehicle, updateVehicle } = useVehicleStore();

  const isEdit = !!id;

  useEffect(() => {
    if (isEdit) {
      setFetching(true);
      vehicleApi
        .getById(id)
        .then(({ data }) => {
          form.setFieldsValue(data.data);
          setVehicleType(data.data.vehicle_type);
        })
        .catch(() => message.error(t('common.error')))
        .finally(() => setFetching(false));
    }
  }, [id]);

  const onFinish = async (values: CreateVehicleRequest) => {
    setLoading(true);
    try {
      if (isEdit) {
        const { data } = await vehicleApi.update(id, values);
        updateVehicle(data.data);
        message.success(t('common.success'));
      } else {
        const { data } = await vehicleApi.create(values);
        addVehicle(data.data);
        message.success(t('common.success'));
      }
      navigate('/vehicles');
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
        <h2>{isEdit ? t('vehicle.editVehicle') : t('vehicle.addVehicle')}</h2>
      </div>

      <Card style={{ maxWidth: 600 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            vehicle_type: 'car',
            fuel_type: 'gasoline',
            is_default: false,
          }}
        >
          <Form.Item
            name="name"
            label={t('vehicle.name')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input placeholder={t('vehicle.name')} />
          </Form.Item>

          <Form.Item
            name="vehicle_type"
            label={t('vehicleType.car')}
            rules={[{ required: true }]}
          >
            <Select
              onChange={(v: VehicleType) => setVehicleType(v)}
              options={VEHICLE_TYPES.map((item) => ({
                value: item.value,
                label: t(item.label),
              }))}
            />
          </Form.Item>

          <Space style={{ width: '100%' }} size="middle">
            <Form.Item
              name="brand"
              label={t('vehicle.brand')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <Input placeholder={t('vehicle.brand')} />
            </Form.Item>

            <Form.Item
              name="model"
              label={t('vehicle.model')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <Input placeholder={t('vehicle.model')} />
            </Form.Item>
          </Space>

          <Space style={{ width: '100%' }} size="middle">
            <Form.Item
              name="year"
              label={t('vehicle.year')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <InputNumber
                min={1900}
                max={new Date().getFullYear() + 1}
                style={{ width: '100%' }}
              />
            </Form.Item>

            <Form.Item
              name="fuel_type"
              label={t('vehicle.fuelType')}
              rules={[{ required: true }]}
              style={{ flex: 1 }}
            >
              <Select
                options={FUEL_TYPES.map((item) => ({
                  value: item.value,
                  label: t(item.label),
                }))}
              />
            </Form.Item>
          </Space>

          <Space style={{ width: '100%' }} size="middle">
            <Form.Item
              name="tank_capacity"
              label={t('vehicle.tankCapacity')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <InputNumber min={1} max={500} style={{ width: '100%' }} suffix="L" />
            </Form.Item>

            {vehicleType === 'motorcycle' && (
              <Form.Item
                name="engine_cc"
                label={t('vehicle.engineCc')}
                style={{ flex: 1 }}
              >
                <InputNumber min={50} max={3000} style={{ width: '100%' }} suffix="cc" />
              </Form.Item>
            )}
          </Space>

          <Form.Item name="is_default" label={t('vehicle.setDefault')} valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {t('common.save')}
              </Button>
              <Button onClick={() => navigate('/vehicles')}>
                {t('common.cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
