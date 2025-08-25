import React from 'react';
import {
  Box,
  Container,
  Typography,
  Button,
  Card,
  CardContent,
  Paper,
} from '@mui/material';
import { Grid2 } from '@mui/material';
import {
  BusinessCenter,
  TrendingUp,
  Security,
  Speed,
  AccountBalance,
  Assessment,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const HomePage: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();

  const features = [
    {
      icon: <Speed />,
      title: 'Fast Financing',
      description: 'Get your invoices financed within 24-48 hours with our streamlined process.',
    },
    {
      icon: <Security />,
      title: 'Blockchain Security',
      description: 'All transactions are secured on the blockchain for maximum transparency.',
    },
    {
      icon: <Assessment />,
      title: 'AI Risk Assessment',
      description: 'Our AI-powered system provides accurate risk scoring for better decisions.',
    },
    {
      icon: <AccountBalance />,
      title: 'Competitive Rates',
      description: 'Access financing at competitive interest rates through our marketplace.',
    },
  ];

  const stats = [
    { label: 'Total Financed', value: '$2.5M+', color: 'primary' },
    { label: 'Active SMEs', value: '450+', color: 'secondary' },
    { label: 'Investors', value: '125+', color: 'success' },
    { label: 'Success Rate', value: '96%', color: 'info' },
  ];

  return (
    <Box>
      {/* Hero Section */}
      <Box
        sx={{
          background: 'linear-gradient(135deg, #1976d2 0%, #1565c0 100%)',
          color: 'white',
          py: 10,
          textAlign: 'center',
        }}
      >
        <Container maxWidth="lg">
          <Typography variant="h2" component="h1" gutterBottom>
            AI-Enabled Invoice Financing Platform
          </Typography>
          <Typography variant="h5" component="p" sx={{ mb: 4, opacity: 0.9 }}>
            Connecting SMEs with investors through blockchain-powered invoice financing
          </Typography>
          
          {!user ? (
            <Box sx={{ display: 'flex', gap: 2, justifyContent: 'center' }}>
              <Button
                variant="contained"
                size="large"
                color="secondary"
                onClick={() => navigate('/register')}
                sx={{ px: 4, py: 1.5 }}
              >
                Get Started
              </Button>
              <Button
                variant="outlined"
                size="large"
                sx={{ px: 4, py: 1.5, borderColor: 'white', color: 'white' }}
                onClick={() => navigate('/marketplace')}
              >
                View Marketplace
              </Button>
            </Box>
          ) : (
            <Button
              variant="contained"
              size="large"
              color="secondary"
              startIcon={<BusinessCenter />}
              onClick={() => {
                const dashboardPath = user.role === 'sme' ? '/sme-dashboard' :
                                   user.role === 'investor' ? '/investor-dashboard' :
                                   '/admin-dashboard';
                navigate(dashboardPath);
              }}
              sx={{ px: 4, py: 1.5 }}
            >
              Go to Dashboard
            </Button>
          )}
        </Container>
      </Box>

      {/* Stats Section */}
      <Container maxWidth="lg" sx={{ py: 6 }}>
        <Grid2 container spacing={3}>
          {stats.map((stat, index) => (
            <Grid2 xs={6} md={3} key={index}>
              <Paper
                elevation={2}
                sx={{
                  p: 3,
                  textAlign: 'center',
                  background: 'linear-gradient(135deg, #f5f5f5 0%, #ffffff 100%)',
                }}
              >
                <Typography variant="h3" color="primary" fontWeight="bold">
                  {stat.value}
                </Typography>
                <Typography variant="body1" color="text.secondary">
                  {stat.label}
                </Typography>
              </Paper>
            </Grid2>
          ))}
        </Grid2>
      </Container>

      {/* Features Section */}
      <Box sx={{ bgcolor: 'background.default', py: 8 }}>
        <Container maxWidth="lg">
          <Typography
            variant="h3"
            component="h2"
            textAlign="center"
            gutterBottom
            color="primary"
          >
            Why Choose Our Platform?
          </Typography>
          <Typography
            variant="h6"
            textAlign="center"
            color="text.secondary"
            sx={{ mb: 6 }}
          >
            Experience the future of invoice financing with our cutting-edge features
          </Typography>

          <Grid2 container spacing={4}>
            {features.map((feature, index) => (
              <Grid2 xs={12} sm={6} md={3} key={index}>
                <Card
                  sx={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                    transition: 'transform 0.2s',
                    '&:hover': {
                      transform: 'translateY(-4px)',
                      boxShadow: 4,
                    },
                  }}
                >
                  <CardContent sx={{ flexGrow: 1, textAlign: 'center', pt: 4 }}>
                    <Box
                      sx={{
                        mb: 2,
                        color: 'primary.main',
                        '& svg': { fontSize: 48 },
                      }}
                    >
                      {feature.icon}
                    </Box>
                    <Typography variant="h6" component="h3" gutterBottom>
                      {feature.title}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {feature.description}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid2>
            ))}
          </Grid2>
        </Container>
      </Box>

      {/* How It Works Section */}
      <Container maxWidth="lg" sx={{ py: 8 }}>
        <Typography
          variant="h3"
          component="h2"
          textAlign="center"
          gutterBottom
          color="primary"
        >
          How It Works
        </Typography>

        <Grid2 container spacing={4} sx={{ mt: 4 }}>
          <Grid2 xs={12} md={6}>
            <Card sx={{ p: 4, height: '100%' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                <BusinessCenter sx={{ fontSize: 40, color: 'primary.main', mr: 2 }} />
                <Typography variant="h4" component="h3" color="primary">
                  For SMEs
                </Typography>
              </Box>
              <Box component="ol" sx={{ pl: 2 }}>
                <Typography component="li" variant="body1" sx={{ mb: 2 }}>
                  Upload your invoices to our secure platform
                </Typography>
                <Typography component="li" variant="body1" sx={{ mb: 2 }}>
                  Our AI system verifies and assesses risk automatically
                </Typography>
                <Typography component="li" variant="body1" sx={{ mb: 2 }}>
                  Create financing requests with competitive rates
                </Typography>
                <Typography component="li" variant="body1">
                  Receive funds within 24-48 hours upon approval
                </Typography>
              </Box>
            </Card>
          </Grid2>

          <Grid2 xs={12} md={6}>
            <Card sx={{ p: 4, height: '100%' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                <TrendingUp sx={{ fontSize: 40, color: 'secondary.main', mr: 2 }} />
                <Typography variant="h4" component="h3" color="secondary">
                  For Investors
                </Typography>
              </Box>
              <Box component="ol" sx={{ pl: 2 }}>
                <Typography component="li" variant="body1" sx={{ mb: 2 }}>
                  Browse verified investment opportunities
                </Typography>
                <Typography component="li" variant="body1" sx={{ mb: 2 }}>
                  Review AI-powered risk assessments and returns
                </Typography>
                <Typography component="li" variant="body1" sx={{ mb: 2 }}>
                  Invest in diversified invoice portfolios
                </Typography>
                <Typography component="li" variant="body1">
                  Earn competitive returns with transparency
                </Typography>
              </Box>
            </Card>
          </Grid2>
        </Grid2>
      </Container>

      {/* CTA Section */}
      {!user && (
        <Box
          sx={{
            bgcolor: 'primary.main',
            color: 'white',
            py: 8,
            textAlign: 'center',
          }}
        >
          <Container maxWidth="md">
            <Typography variant="h4" component="h2" gutterBottom>
              Ready to Get Started?
            </Typography>
            <Typography variant="h6" sx={{ mb: 4, opacity: 0.9 }}>
              Join thousands of SMEs and investors already using our platform
            </Typography>
            <Box sx={{ display: 'flex', gap: 2, justifyContent: 'center' }}>
              <Button
                variant="contained"
                size="large"
                color="secondary"
                onClick={() => navigate('/register')}
                sx={{ px: 4, py: 1.5 }}
              >
                Register Now
              </Button>
              <Button
                variant="outlined"
                size="large"
                sx={{ px: 4, py: 1.5, borderColor: 'white', color: 'white' }}
                onClick={() => navigate('/login')}
              >
                Sign In
              </Button>
            </Box>
          </Container>
        </Box>
      )}
    </Box>
  );
};

export default HomePage;
