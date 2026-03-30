import { Typography, Card, Space, Divider } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const { Title, Paragraph, Text } = Typography;

export default function TermsOfServicePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const handleBack = () => {
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
          <Title level={2}>{t('legal.termsOfService')}</Title>
          <Text type="secondary">{t('legal.lastUpdated')}: 2026-03-30</Text>

          <Divider />

          <Title level={4}>{t('terms.acceptanceTitle')}</Title>
          <Paragraph>{t('terms.acceptanceContent')}</Paragraph>

          <Title level={4}>{t('terms.serviceTitle')}</Title>
          <Paragraph>{t('terms.serviceContent')}</Paragraph>
          <Paragraph>
            <ul>
              <li>{t('terms.serviceTrack')}</li>
              <li>{t('terms.serviceAnalyze')}</li>
              <li>{t('terms.serviceManage')}</li>
              <li>{t('terms.serviceExport')}</li>
            </ul>
          </Paragraph>

          <Title level={4}>{t('terms.accountTitle')}</Title>
          <Paragraph>{t('terms.accountContent')}</Paragraph>

          <Title level={4}>{t('terms.userResponsibilityTitle')}</Title>
          <Paragraph>{t('terms.userResponsibilityContent')}</Paragraph>
          <Paragraph>
            <ul>
              <li>{t('terms.responsibilityAccurate')}</li>
              <li>{t('terms.responsibilitySecure')}</li>
              <li>{t('terms.responsibilityLawful')}</li>
            </ul>
          </Paragraph>

          <Title level={4}>{t('terms.disclaimerTitle')}</Title>
          <Paragraph>{t('terms.disclaimerContent')}</Paragraph>

          <Title level={4}>{t('terms.terminationTitle')}</Title>
          <Paragraph>{t('terms.terminationContent')}</Paragraph>

          <Title level={4}>{t('terms.changesTitle')}</Title>
          <Paragraph>{t('terms.changesContent')}</Paragraph>

          <Title level={4}>{t('terms.contactTitle')}</Title>
          <Paragraph>{t('terms.contactContent')}</Paragraph>
        </Typography>
      </Card>
    </div>
  );
}
