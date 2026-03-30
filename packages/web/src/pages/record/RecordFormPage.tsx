import { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  InputNumber,
  DatePicker,
  Select,
  Switch,
  Button,
  Space,
  message,
  Spin,
  AutoComplete,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  fuelRecordApi,
  vehicleApi,
  useAuthStore,
  CURRENCIES,
  FUEL_GRADES,
  getFuelGradesByLocale,
  isElectricVehicle,
} from '@gastrack/shared';
import type { CreateFuelRecordRequest, Vehicle } from '@gastrack/shared';
import dayjs from 'dayjs';

/**
 * 自动计算逻辑（改进版）：
 * - 加油量 × 单价 = 总费用
 * - 追踪用户最后手动编辑的两个字段（editStack），以这两个为准自动计算第三个
 * - 例：先填加油量再填单价 → 自动算总费用；接着改总费用 → 以单价+总费用为准算加油量
 */

type CalcField = 'fuel_amount' | 'unit_price' | 'total_cost';

export default function RecordFormPage() {
  const { t, i18n: i18nInstance } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId: urlVehicleId, recordId } = useParams<{
    vehicleId: string;
    recordId: string;
  }>();
  const [searchParams] = useSearchParams();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [isEv, setIsEv] = useState(false);
  const [stationSuggestions, setStationSuggestions] = useState<string[]>([]);
  const [stationSearch, setStationSearch] = useState('');
  const user = useAuthStore((s) => s.user);

  // 车辆选择器状态（当没有 URL vehicleId 时启用）
  const [allVehicles, setAllVehicles] = useState<Vehicle[]>([]);
  const [selectedVehicleId, setSelectedVehicleId] = useState<string | undefined>(
    urlVehicleId || searchParams.get('vehicleId') || undefined,
  );
  const [vehiclesLoading, setVehiclesLoading] = useState(false);

  // 有效的 vehicleId（来自 URL 参数或手动选择）
  const vehicleId = urlVehicleId || selectedVehicleId;

  // 是否需要显示车辆选择器（无 URL vehicleId 参数时）
  const showVehicleSelector = !urlVehicleId;

  /**
   * 编辑栈：记录用户最近手动编辑的字段顺序（最多保留2个）
   * 栈顶（末尾）为最新编辑的字段
   * 例：[unit_price, fuel_amount] 表示最近编辑了单价再编辑了加油量 → 自动计算总费用
   */
  const editStackRef = useRef<CalcField[]>([]);

  const isEdit = !!recordId;

  // 根据用户设置推断默认单位
  const isImperial = user?.unit_system === 'imperial';
  const defaultFuelUnit = isImperial ? 'gal' : 'L';
  const defaultDistanceUnit = isImperial ? 'mi' : 'km';
  const defaultCurrency = user?.currency_code || 'CNY';

  // 加油站建议：根据输入文字模糊过滤
  const stationOptions = useMemo(() => {
    if (!stationSearch) {
      return stationSuggestions.map((name) => ({ value: name, label: name }));
    }
    const lower = stationSearch.toLowerCase();
    return stationSuggestions
      .filter((name) => name.toLowerCase().includes(lower))
      .map((name) => ({ value: name, label: name }));
  }, [stationSuggestions, stationSearch]);

  // 加载车辆列表（仅当需要车辆选择器时）
  useEffect(() => {
    if (showVehicleSelector) {
      setVehiclesLoading(true);
      vehicleApi
        .list({ include_shared: true })
        .then(({ data }) => {
          const vehicles = data.data || [];
          setAllVehicles(vehicles);
          // 如果初始 vehicleId 不在列表中（如被取消共享），清空选择
          if (selectedVehicleId && !vehicles.some((v: Vehicle) => v.id === selectedVehicleId)) {
            setSelectedVehicleId(undefined);
          }
        })
        .catch(() => { /* ignore */ })
        .finally(() => setVehiclesLoading(false));
    }
  }, [showVehicleSelector]);

  // 构建车辆选择器选项（分组："我的车辆" / "共享车辆（来自XX群组）"）
  const vehicleOptions = useMemo(() => {
    const myVehicles = allVehicles.filter((v) => !v.shared_from_group_id);
    const sharedVehicles = allVehicles.filter((v) => !!v.shared_from_group_id);

    // 按群组名分组
    const sharedByGroup = new Map<string, { groupName: string; vehicles: Vehicle[] }>();
    for (const v of sharedVehicles) {
      const groupId = v.shared_from_group_id!;
      if (!sharedByGroup.has(groupId)) {
        sharedByGroup.set(groupId, { groupName: v.shared_from_group_name || groupId, vehicles: [] });
      }
      sharedByGroup.get(groupId)!.vehicles.push(v);
    }

    const options: { label: string; options: { value: string; label: string }[] }[] = [];

    if (myVehicles.length > 0) {
      options.push({
        label: t('fuelRecord.myVehicles'),
        options: myVehicles.map((v) => ({
          value: v.id,
          label: v.name,
        })),
      });
    }

    for (const [, group] of sharedByGroup) {
      options.push({
        label: t('fuelRecord.sharedVehicles', { groupName: group.groupName }),
        options: group.vehicles.map((v) => ({
          value: v.id,
          label: v.name,
        })),
      });
    }

    return options;
  }, [allVehicles, t]);

  // 获取车辆信息以判断是否为电动车
  useEffect(() => {
    if (vehicleId) {
      vehicleApi.getById(vehicleId).then(({ data }) => {
        const vehicle = data.data;
        const ev = isElectricVehicle(vehicle.fuel_type);
        setIsEv(ev);
        // 电动车默认能量单位为 kWh
        if (ev && !isEdit) {
          form.setFieldValue('fuel_unit', 'kWh');
        }
        // 自动带入车辆的默认燃油标号（仅新建时）
        if (!ev && !isEdit && vehicle.fuel_grade) {
          form.setFieldValue('fuel_grade', vehicle.fuel_grade);
        }
      }).catch(() => { /* ignore */ });

      // 加载加油站/充电站建议列表
      fuelRecordApi.getStationSuggestions(vehicleId).then(({ data }) => {
        setStationSuggestions(data.data || []);
      }).catch(() => { /* ignore */ });
    }
  }, [vehicleId]);

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

  // 自动计算：追踪最后编辑的两个字段，以它们为准计算第三个
  const autoCalc = useCallback(
    (changedField: CalcField) => {
      // 1. 更新编辑栈：把当前字段推入栈顶，去重后保留最近2个
      const stack = editStackRef.current.filter((f) => f !== changedField);
      stack.push(changedField);
      if (stack.length > 2) stack.shift();
      editStackRef.current = stack;

      // 2. 读取三个字段当前值
      const amount = form.getFieldValue('fuel_amount') as number | undefined;
      const price = form.getFieldValue('unit_price') as number | undefined;
      const total = form.getFieldValue('total_cost') as number | undefined;

      // 3. 需要两个有值的字段才能计算
      if (stack.length < 2) return;

      // 4. 确定要自动计算的第三个字段（不在编辑栈中的那个）
      const allFields: CalcField[] = ['fuel_amount', 'unit_price', 'total_cost'];
      const computedField = allFields.find((f) => !stack.includes(f))!;

      // 5. 根据两个已知字段计算第三个
      const round2 = (n: number) => Math.round(n * 100) / 100;

      if (computedField === 'total_cost') {
        // 总费用 = 加油量 × 单价
        if (amount && price) {
          form.setFieldValue('total_cost', round2(amount * price));
        }
      } else if (computedField === 'unit_price') {
        // 单价 = 总费用 ÷ 加油量
        if (total && amount) {
          form.setFieldValue('unit_price', round2(total / amount));
        }
      } else if (computedField === 'fuel_amount') {
        // 加油量 = 总费用 ÷ 单价
        if (total && price) {
          form.setFieldValue('fuel_amount', round2(total / price));
        }
      }
    },
    [form],
  );

  // 车辆选择变化时的处理
  const handleVehicleChange = useCallback((value: string) => {
    setSelectedVehicleId(value);
    // 重置电车状态和燃油标号
    const vehicle = allVehicles.find((v) => v.id === value);
    if (vehicle) {
      const ev = isElectricVehicle(vehicle.fuel_type);
      setIsEv(ev);
      if (ev) {
        form.setFieldValue('fuel_unit', 'kWh');
      } else {
        form.setFieldValue('fuel_unit', defaultFuelUnit);
        if (vehicle.fuel_grade) {
          form.setFieldValue('fuel_grade', vehicle.fuel_grade);
        }
      }
    }
  }, [allVehicles, form, defaultFuelUnit]);

  const onFinish = async (values: CreateFuelRecordRequest & { refuel_date: dayjs.Dayjs }) => {
    if (!vehicleId) {
      message.error(t('fuelRecord.selectVehiclePlaceholder'));
      return;
    }

    setLoading(true);
    const payload: CreateFuelRecordRequest = {
      ...values,
      refuel_date: values.refuel_date.format('YYYY-MM-DDTHH:mm:ssZ'),
      currency_code: values.currency_code || defaultCurrency,
      fuel_unit: values.fuel_unit || defaultFuelUnit,
      distance_unit: values.distance_unit || defaultDistanceUnit,
    };

    try {
      if (isEdit) {
        await fuelRecordApi.update(vehicleId!, recordId!, payload);
      } else {
        await fuelRecordApi.create(vehicleId!, payload);
      }
      message.success(t('common.success'));
      // 如果是通过 URL vehicleId 进入的，返回该车辆的记录页
      // 否则返回到之前的页面（如群组页）
      if (urlVehicleId) {
        navigate(`/vehicles/${vehicleId}/records`);
      } else {
        navigate(-1);
      }
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
            onClick={() => {
              if (urlVehicleId) {
                navigate(`/vehicles/${urlVehicleId}/records`);
              } else {
                navigate(-1);
              }
            }}
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
            fuel_unit: defaultFuelUnit,
            currency_code: defaultCurrency,
            distance_unit: defaultDistanceUnit,
          }}
        >
          {/* 车辆选择器（仅当没有 URL vehicleId 时显示） */}
          {showVehicleSelector && (
            <Form.Item
              label={t('fuelRecord.selectVehicle')}
              required
            >
              <Select
                value={vehiclesLoading ? undefined : selectedVehicleId}
                onChange={handleVehicleChange}
                placeholder={t('fuelRecord.selectVehiclePlaceholder')}
                loading={vehiclesLoading}
                options={vehicleOptions}
                showSearch
                optionFilterProp="label"
                style={{ width: '100%' }}
              />
            </Form.Item>
          )}
          <Form.Item
            name="refuel_date"
            label={isEv ? t('fuelRecord.chargingDate') : t('fuelRecord.fuelDate')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <DatePicker
              showTime
              style={{ width: '100%' }}
              format="YYYY-MM-DD HH:mm"
            />
          </Form.Item>

          <Form.Item name="station_name" label={isEv ? t('fuelRecord.chargingStation') : t('fuelRecord.station')}>
            <AutoComplete
              options={stationOptions}
              onSearch={setStationSearch}
              placeholder={isEv ? t('fuelRecord.chargingStationPlaceholder') : t('fuelRecord.stationPlaceholder')}
              allowClear
              filterOption={false}
            />
          </Form.Item>

          {/* 加油量/充电量（单位跟随用户设置） */}
          <Form.Item
            name="fuel_amount"
            label={isEv ? t('fuelRecord.chargingAmount') : t('fuelRecord.fuelAmount')}
          >
            <InputNumber
              min={0.01}
              step={0.01}
              style={{ width: '100%' }}
              suffix={isEv ? 'kWh' : defaultFuelUnit}
              onChange={() => autoCalc('fuel_amount')}
            />
          </Form.Item>
          <Form.Item name="fuel_unit" hidden><Input /></Form.Item>

          {/* 单价（货币跟随用户设置） */}
          <Form.Item
            name="unit_price"
            label={t('fuelRecord.pricePerUnit')}
          >
            <InputNumber
              min={0.01}
              step={0.01}
              style={{ width: '100%' }}
              prefix={CURRENCIES.find((c) => c.value === defaultCurrency)?.symbol || defaultCurrency}
              suffix={`/${isEv ? 'kWh' : defaultFuelUnit}`}
              onChange={() => autoCalc('unit_price')}
            />
          </Form.Item>
          <Form.Item name="currency_code" hidden><Input /></Form.Item>

          {/* 总费用 */}
          <Form.Item
            name="total_cost"
            label={t('fuelRecord.totalCost')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <InputNumber
              min={0}
              step={0.01}
              style={{ width: '100%' }}
              prefix={CURRENCIES.find((c) => c.value === defaultCurrency)?.symbol || defaultCurrency}
              onChange={() => autoCalc('total_cost')}
            />
          </Form.Item>

          {/* 里程（单位跟随用户设置） */}
          <Form.Item
            name="odometer"
            label={t('fuelRecord.odometer')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <InputNumber min={0} step={1} style={{ width: '100%' }} suffix={defaultDistanceUnit} />
          </Form.Item>
          <Form.Item name="distance_unit" hidden><Input /></Form.Item>

          {/* 燃油标号（仅燃油车显示）——按当前语言优先显示本地区标号 */}
          {!isEv && (() => {
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

          <Form.Item
            name="is_full_tank"
            label={isEv ? t('fuelRecord.isFullCharge') : t('fuelRecord.isFullTank')}
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
                onClick={() => {
                  if (urlVehicleId) {
                    navigate(`/vehicles/${urlVehicleId}/records`);
                  } else {
                    navigate(-1);
                  }
                }}
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
