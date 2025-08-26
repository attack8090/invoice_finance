import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  Card,
  CardContent,
  Grid,
  Chip,
  Alert,
  CircularProgress,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  LinearProgress
} from '@mui/material';
import {
  TrendingUp,
  AttachMoney,
  Assessment,
  Timeline,
  Add,
  AccountBalance,
  ShowChart
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

interface Investment {
  id: string;
  financing_request_id: string;
  amount: number;
  expected_return: number;
  actual_return: number;
  status: 'pending' | 'active' | 'completed' | 'defaulted';
  investment_date: string;
  maturity_date: string;
  financing_request: {
    interest_rate: number;
    risk_level: string;
    invoice: {
      invoice_number: string;
      customer_name: string;
      user: {
        company_name: string;
      };
    };
  };
}

interface InvestmentOpportunity {
  id: string;
  requested_amount: number;
  interest_rate: number;
  risk_level: string;
  status: string;
  created_at: string;
  invoice: {
    invoice_number: string;
    customer_name: string;
    invoice_amount: number;
    due_date: string;
    user: {
      company_name: string;
      credit_score: number;
    };
  };
}

interface DashboardStats {
  totalInvested: number;
  totalReturns: number;
  activeInvestments: number;
  completedInvestments: number;
  averageReturn: number;
  portfolioValue: number;
}

const InvestorDashboard: React.FC = () => {
  const [investments, setInvestments] = useState<Investment[]>([]);
  const [opportunities, setOpportunities] = useState<InvestmentOpportunity[]>([]);
  const [stats, setStats] = useState<DashboardStats>({
    totalInvested: 0,
    totalReturns: 0,
    activeInvestments: 0,
    completedInvestments: 0,
    averageReturn: 0,
    portfolioValue: 0
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { user } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      const [investmentsResponse, opportunitiesResponse] = await Promise.all([
        axios.get('/financing/investments'),
        axios.get('/financing/opportunities')
      ]);

      const investmentsData = investmentsResponse.data || [];
      const opportunitiesData = opportunitiesResponse.data || [];

      setInvestments(investmentsData.slice(0, 5)); // Show recent 5
      setOpportunities(opportunitiesData.slice(0, 5)); // Show recent 5

      // Calculate stats
      const totalInvested = investmentsData.reduce((sum: number, inv: Investment) => sum + inv.amount, 0);
      const totalReturns = investmentsData.reduce((sum: number, inv: Investment) => sum + inv.actual_return, 0);
      const activeInvestments = investmentsData.filter((inv: Investment) => inv.status === 'active').length;
      const completedInvestments = investmentsData.filter((inv: Investment) => inv.status === 'completed').length;
      const averageReturn = investmentsData.length > 0 ? 
        (totalReturns / totalInvested) * 100 : 0;
      const portfolioValue = totalInvested + totalReturns;

      setStats({
        totalInvested,
        totalReturns,
        activeInvestments,
        completedInvestments,
        averageReturn,
        portfolioValue
      });
    } catch (err: any) {
      setError('Failed to fetch dashboard data');
      console.error('Error fetching dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getStatusColor = (status: string) => {
    const statusColors: Record<string, 'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning'> = {
      pending: 'warning',
      approved: 'info',
      active: 'primary',
      completed: 'success',
      defaulted: 'error',
      funded: 'success'
    };
    return statusColors[status] || 'default';
  };

  const getRiskColor = (riskLevel: string) => {
    switch (riskLevel) {
      case 'low': return 'success';
      case 'medium': return 'warning';
      case 'high': return 'error';
      default: return 'default';
    }
  };

  if (loading) {
    return (
      <Container maxWidth="lg">
        <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '50vh' }}>
          <CircularProgress />
        </Box>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Investor Dashboard
        </Typography>
        <Typography variant="body1" color="textSecondary" sx={{ mb: 4 }}>
          Welcome back, {user?.company_name || user?.email}
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
            {error}
          </Alert>
        )}

        {/* Key Metrics */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <AccountBalance sx={{ fontSize: 40, color: 'primary.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Portfolio Value
                    </Typography>
                    <Typography variant="h5">
                      {formatCurrency(stats.portfolioValue)}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <AttachMoney sx={{ fontSize: 40, color: 'info.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Total Invested
                    </Typography>
                    <Typography variant="h5">
                      {formatCurrency(stats.totalInvested)}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <TrendingUp sx={{ fontSize: 40, color: 'success.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Total Returns
                    </Typography>
                    <Typography variant="h5">
                      {formatCurrency(stats.totalReturns)}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Assessment sx={{ fontSize: 40, color: 'secondary.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Active Investments
                    </Typography>
                    <Typography variant="h5">
                      {stats.activeInvestments}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Timeline sx={{ fontSize: 40, color: 'warning.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Completed Investments
                    </Typography>
                    <Typography variant="h5">
                      {stats.completedInvestments}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <ShowChart sx={{ fontSize: 40, color: 'success.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Average Return
                    </Typography>
                    <Typography variant="h5">
                      {stats.averageReturn.toFixed(1)}%
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        <Grid container spacing={3}>
          {/* Recent Investments */}
          <Grid item xs={12} lg={6}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6" component="h2">
                    Recent Investments
                  </Typography>
                  <Button
                    startIcon={<Assessment />}
                    variant="contained"
                    size="small"
                    onClick={() => navigate('/investments')}
                  >
                    View All
                  </Button>
                </Box>
                <TableContainer>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Company</TableCell>
                        <TableCell>Amount</TableCell>
                        <TableCell>Return</TableCell>
                        <TableCell>Status</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {investments.map((investment) => (
                        <TableRow key={investment.id}>
                          <TableCell>
                            <Box>
                              <Typography variant="body2" fontWeight="medium">
                                {investment.financing_request.invoice.user.company_name}
                              </Typography>
                              <Typography variant="caption" color="textSecondary">
                                {investment.financing_request.invoice.invoice_number}
                              </Typography>
                            </Box>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2">
                              {formatCurrency(investment.amount)}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2" color="success.main">
                              {formatCurrency(investment.expected_return)}
                            </Typography>
                            <Typography variant="caption" color="textSecondary">
                              ({investment.financing_request.interest_rate}%)
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={investment.status.charAt(0).toUpperCase() + investment.status.slice(1)}
                              color={getStatusColor(investment.status) as any}
                              size="small"
                            />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
                {investments.length === 0 && (
                  <Box sx={{ textAlign: 'center', py: 3 }}>
                    <Typography variant="body2" color="textSecondary">
                      No investments found. Explore the marketplace to start investing.
                    </Typography>
                  </Box>
                )}
              </CardContent>
            </Card>
          </Grid>

          {/* New Investment Opportunities */}
          <Grid item xs={12} lg={6}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6" component="h2">
                    New Opportunities
                  </Typography>
                  <Button
                    startIcon={<Add />}
                    variant="contained"
                    size="small"
                    onClick={() => navigate('/marketplace')}
                  >
                    Browse All
                  </Button>
                </Box>
                <TableContainer>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Company</TableCell>
                        <TableCell>Amount</TableCell>
                        <TableCell>Return</TableCell>
                        <TableCell>Risk</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {opportunities.map((opportunity) => (
                        <TableRow key={opportunity.id}>
                          <TableCell>
                            <Box>
                              <Typography variant="body2" fontWeight="medium">
                                {opportunity.invoice.user.company_name}
                              </Typography>
                              <Typography variant="caption" color="textSecondary">
                                Score: {opportunity.invoice.user.credit_score}
                              </Typography>
                            </Box>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2">
                              {formatCurrency(opportunity.requested_amount)}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2" color="success.main">
                              {opportunity.interest_rate}%
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={opportunity.risk_level.toUpperCase()}
                              color={getRiskColor(opportunity.risk_level) as any}
                              size="small"
                            />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
                {opportunities.length === 0 && (
                  <Box sx={{ textAlign: 'center', py: 3 }}>
                    <Typography variant="body2" color="textSecondary">
                      No new opportunities available. Check back later.
                    </Typography>
                  </Box>
                )}
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </Box>
    </Container>
  );
};

export default InvestorDashboard;
