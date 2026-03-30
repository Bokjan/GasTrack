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
  Row,
  Col,
  message,
  Spin,
} from 'antd';
import { useTranslation } from 'react-i18next';
import {
  vehicleApi,
  useVehicleStore,
  useAuthStore,
  FUEL_TYPES,
  VEHICLE_TYPES,
  FUEL_UNITS,
  ENERGY_UNITS,
  FUEL_GRADES,
  getFuelGradesByLocale,
  isElectricVehicle,
  hasEngineCC,
} from '@gastrack/shared';
import type { CreateVehicleRequest, FuelType } from '@gastrack/shared';

export default function VehicleFormPage() {
  const { t, i18n: i18nInstance } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [fuelType, setFuelType] = useState<FuelType>('gasoline');
  const { addVehicle, updateVehicle } = useVehicleStore();
  const user = useAuthStore((s) => s.user);

  const isImperial = user?.unit_system === 'imperial';
  const defaultCapacityUnit = isImperial ? 'gal' : 'L';

  const isEdit = !!id;

  useEffect(() => {
    if (isEdit) {
      setFetching(true);
      vehicleApi
        .getById(id)
        .then(({ data }) => {
          form.setFieldsValue(data.data);
          setFuelType(data.data.fuel_type);
        })
        .catch(() => message.error(t('common.error')))
        .finally(() => setFetching(false));
    }
  }, [id]);

  const onFinish = async (values: CreateVehicleRequest & { tank_capacity_unit?: string }) => {
    setLoading(true);
    try {
      // tank_capacity_unit 仅用于前端 UI 展示，后端不接受该字段
      const { tank_capacity_unit: _unused, ...payload } = values;
      if (isEdit) {
        const { data } = await vehicleApi.update(id, payload);
        updateVehicle(data.data);
        message.success(t('common.success'));
      } else {
        const { data } = await vehicleApi.create(payload);
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
            tank_capacity_unit: defaultCapacityUnit,
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
            label={t('vehicle.vehicleType')}
            rules={[{ required: true }]}
          >
            <Select
              options={VEHICLE_TYPES.map((item) => ({
                value: item.value,
                label: t(item.label),
              }))}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name="brand"
                label={t('vehicle.brand')}
                rules={[{ required: true, message: t('common.required') }]}
              >
                <Input placeholder={t('vehicle.brand')} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="model"
                label={t('vehicle.model')}
                rules={[{ required: true, message: t('common.required') }]}
              >
                <Input placeholder={t('vehicle.model')} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name="year"
                label={t('vehicle.year')}
                rules={[{ required: true, message: t('common.required') }]}
              >
                <InputNumber
                  min={1900}
                  max={new Date().getFullYear() + 1}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="fuel_type"
                label={t('vehicle.fuelType')}
                rules={[{ required: true }]}
              >
                <Select
                  popupMatchSelectWidth={false}
                  onChange={(v: FuelType) => {
                    setFuelType(v);
                    // 切换为电动时，自动将容量单位改为 kWh
                    if (isElectricVehicle(v)) {
                      form.setFieldValue('tank_capacity_unit', 'kWh');
                    } else if (form.getFieldValue('tank_capacity_unit') === 'kWh') {
                      form.setFieldValue('tank_capacity_unit', defaultCapacityUnit);
                    }
                  }}
                  options={FUEL_TYPES.map((item) => ({
                    value: item.value,
                    label: t(item.label),
                  }))}
                />
              </Form.Item>
            </Col>
          </Row>

          {/* 燃油标号（仅燃油车/柴油车/混动车显示） */}
          {!isElectricVehicle(fuelType) && (() => {
            const localGrades = getFuelGradesByLocale(i18nInstance.language);
            const allGrades = FUEL_GRADES;
            const localValues = new Set(localGrades.map((g) => g.value));
            const otherGrades = allGrades.filter((g) => !localValues.has(g.value));

            return (
              <Form.Item name="fuel_grade" label={t('fuelRecord.fuelGrade')}>
                <Select
                  allowClear
                  placeholder={t('fuelRecord.fuelGradePlaceholder')}
                  options={[
                    {
                      label: t('fuelGrade.localGroup'),
                      options: localGrades.map((g) => ({
                        value: g.value,
                        label: t(g.label),
                      })),
                    },
                    {
                      label: t('fuelGrade.otherGroup'),
                      options: otherGrades.map((g) => ({
                        value: g.value,
                        label: t(g.label),
                      })),
                    },
                  ]}
                />
              </Form.Item>
            );
          })()}

          <Row gutter={16}>
            <Col xs={24} sm={14}>
              <Form.Item
                name="tank_capacity"
                label={isElectricVehicle(fuelType) ? t('vehicle.batteryCapacity') : t('vehicle.tankCapacity')}
                rules={[{ required: true, message: t('common.required') }]}
              >
                <InputNumber min={1} max={isElectricVehicle(fuelType) ? 200 : 500} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={10}>
              <Form.Item
                name="tank_capacity_unit"
                label={isElectricVehicle(fuelType) ? t('fuelRecord.energyUnit') : t('fuelRecord.fuelUnit')}
              >
                <Select
                  options={(isElectricVehicle(fuelType) ? ENERGY_UNITS : FUEL_UNITS).map((u) => ({
                    value: u.value,
                    label: t(u.label),
                  }))}
                />
              </Form.Item>
            </Col>
          </Row>

          {hasEngineCC(fuelType) && (
            <Form.Item
              name="engine_cc"
              label={t('vehicle.engineCc')}
            >
              <InputNumber min={50} max={10000} style={{ width: '100%', maxWidth: 300 }} suffix="cc" />
            </Form.Item>
          )}

          <Form.Item
            name="license_plate"
            label={t('vehicle.licensePlate')}
          >
            <Input
              placeholder={t('vehicle.licensePlatePlaceholder')}
              maxLength={20}
              style={{ maxWidth: 300 }}
            />
          </Form.Item>

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
