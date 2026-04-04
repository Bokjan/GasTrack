import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Button, List, Tag, Space, Popconfirm, message, Empty, Tooltip } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  StarOutlined,
  StarFilled,
  WalletOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useVehicleStore, vehicleApi, FUEL_GRADES } from '@gastrack/shared';

export default function VehicleListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicles, fetchVehicles, removeVehicle, isLoading } = useVehicleStore();

  useEffect(() => {
    fetchVehicles();
  }, []);

  const handleDelete = async (id: string) => {
    try {
      await vehicleApi.delete(id);
      removeVehicle(id);
      message.success(t('common.success'));
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleSetDefault = async (id: string) => {
    try {
      await vehicleApi.update(id, { is_default: true });
      // 重新获取列表以更新默认状态
      fetchVehicles();
      message.success(t('common.success'));
    } catch {
      message.error(t('common.error'));
    }
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>{t('vehicle.title')}</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate('/vehicles/new')}
        >
          {t('vehicle.addVehicle')}
        </Button>
      </div>

      {vehicles.length === 0 && !isLoading ? (
        <Card>
          <Empty description={t('vehicle.noVehicle')}>
            <Button type="primary" onClick={() => navigate('/vehicles/new')}>
              {t('vehicle.addFirst')}
            </Button>
          </Empty>
        </Card>
      ) : (
        <List
          loading={isLoading}
          grid={{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3, xl: 3 }}
          dataSource={vehicles}
          renderItem={(vehicle) => (
            <List.Item>
              <Card
                className="vehicle-card"
                hoverable
                onClick={() => navigate(`/vehicles/${vehicle.id}/records`)}
                actions={[
                  <Button
                    key="default"
                    type="text"
                    icon={vehicle.is_default ? <StarFilled style={{ color: '#faad14' }} /> : <StarOutlined />}
                    onClick={(e) => {
                      e.stopPropagation();
                      if (!vehicle.is_default) handleSetDefault(vehicle.id);
                    }}
                  />,
                  <Tooltip title={t('expense.title')} key="expenses">
                    <Button
                      type="text"
                      icon={<WalletOutlined />}
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(`/vehicles/${vehicle.id}/expenses`);
                      }}
                    />
                  </Tooltip>,
                  <Button
                    key="edit"
                    type="text"
                    icon={<EditOutlined />}
                    onClick={(e) => {
                      e.stopPropagation();
                      navigate(`/vehicles/${vehicle.id}/edit`);
                    }}
                  />,
                  <Popconfirm
                    key="delete"
                    title={t('vehicle.deleteConfirm')}
                    onConfirm={(e) => {
                      e?.stopPropagation();
                      handleDelete(vehicle.id);
                    }}
                    onCancel={(e) => e?.stopPropagation()}
                  >
                    <Button
                      type="text"
                      danger
                      icon={<DeleteOutlined />}
                      onClick={(e) => e.stopPropagation()}
                    />
                  </Popconfirm>,
                ]}
              >
                <Card.Meta
                  avatar={
                    <div style={{ fontSize: 40 }}>
                      {vehicle.vehicle_type === 'motorcycle' ? '🏍️' : '🚗'}
                    </div>
                  }
                  title={
                    <Space>
                      <span>{vehicle.name}</span>
                      {vehicle.is_default && <Tag color="blue">{t('vehicle.default')}</Tag>}
                    </Space>
                  }
                  description={
                    <>
                      <div>{vehicle.brand} {vehicle.model} · {vehicle.year}</div>
                      {vehicle.license_plate && (
                        <div style={{ marginTop: 2, color: 'var(--gt-text-secondary, #666)' }}>
                          {vehicle.license_plate}
                        </div>
                      )}
                      <div style={{ marginTop: 4 }}>
                        <Tag>{t(`fuelType.${vehicle.fuel_type}`)}</Tag>
                        {vehicle.fuel_grade && (() => {
                          const grade = FUEL_GRADES.find((g) => g.value === vehicle.fuel_grade);
                          return <Tag color="green">{grade ? t(grade.label) : vehicle.fuel_grade}</Tag>;
                        })()}
                        <Tag>{t(`vehicleType.${vehicle.vehicle_type}`)}</Tag>
                        {vehicle.engine_cc && <Tag>{vehicle.engine_cc}cc</Tag>}
                      </div>
                    </>
                  }
                />
              </Card>
            </List.Item>
          )}
        />
      )}
    </div>
  );
}
