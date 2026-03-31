import { useEffect } from 'react';
import { Typography, Card, Space, Divider } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const { Title, Paragraph, Text } = Typography;

export default function PrivacyPolicyPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  const handleBack = () => {
    // 如果有浏览器历史记录（非新标签页打开），则返回上一页；否则导航到首页
    if (window.history.length > 1) {
      navigate(-1);
    } else {
      navigate('/');
    }
  };

  return (
    <div style={{ maxWidth: 800, margin: '0 auto', padding: '24px 16px' }}>
      <Space style={{ marginBottom: 16, cursor: 'pointer' }} onClick={handleBack}>
        <ArrowLeftOutlined />
        <Text>{t('common.back')}</Text>
      </Space>

      <Card>
        <Typography>
          <Title level={2}>{t('legal.privacyPolicy')}</Title>
          <Text type="secondary">{t('legal.lastUpdated')}: 2026-03-30</Text>

          <Divider />

          <Title level={4}>{t('privacy.introTitle')}</Title>
          <Paragraph>{t('privacy.introContent')}</Paragraph>

          <Title level={4}>{t('privacy.dataCollectedTitle')}</Title>
          <Paragraph>{t('privacy.dataCollectedContent')}</Paragraph>
          <Paragraph>
            <ul>
              <li>{t('privacy.dataAccount')}</li>
              <li>{t('privacy.dataVehicle')}</li>
              <li>{t('privacy.dataRecords')}</li>
              <li>{t('privacy.dataPreferences')}</li>
            </ul>
          </Paragraph>

          <Title level={4}>{t('privacy.dataUsageTitle')}</Title>
          <Paragraph>{t('privacy.dataUsageContent')}</Paragraph>
          <Paragraph>
            <ul>
              <li>{t('privacy.usageProvideService')}</li>
              <li>{t('privacy.usageAnalytics')}</li>
              <li>{t('privacy.usageImprove')}</li>
            </ul>
          </Paragraph>

          <Title level={4}>{t('privacy.dataStorageTitle')}</Title>
          <Paragraph>{t('privacy.dataStorageContent')}</Paragraph>

          <Title level={4}>{t('privacy.dataSharingTitle')}</Title>
          <Paragraph>{t('privacy.dataSharingContent')}</Paragraph>

          <Title level={4}>{t('privacy.userRightsTitle')}</Title>
          <Paragraph>{t('privacy.userRightsContent')}</Paragraph>
          <Paragraph>
            <ul>
              <li>{t('privacy.rightAccess')}</li>
              <li>{t('privacy.rightExport')}</li>
              <li>{t('privacy.rightDelete')}</li>
              <li>{t('privacy.rightModify')}</li>
            </ul>
          </Paragraph>

          <Title level={4}>{t('privacy.localStorageTitle')}</Title>
          <Paragraph>{t('privacy.localStorageContent')}</Paragraph>
          <Paragraph>
            <ul>
              <li><Text code>access_token</Text> / <Text code>refresh_token</Text> — {t('privacy.storageAuth')}</li>
              <li><Text code>locale</Text> — {t('privacy.storageLocale')}</li>
              <li><Text code>theme_mode</Text> — {t('privacy.storageTheme')}</li>
            </ul>
          </Paragraph>

          <Title level={4}>{t('privacy.contactTitle')}</Title>
          <Paragraph>{t('privacy.contactContent')}</Paragraph>
        </Typography>
      </Card>
    </div>
  );
}
