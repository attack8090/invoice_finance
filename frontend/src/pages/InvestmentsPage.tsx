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
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  LinearProgress,
  Tab,
  Tabs
} from '@mui/material';
import {
  TrendingUp,
  AttachMoney,
  DateRange,
  Assessment
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import axios from 'axios';

interface Investment {
  id: string;
  financing_request_id: string;
  investor_id: string;
  amount: number;
  expected_return: number;
  actual_return: number;
  status: 'pending' | 'active' | 'completed' | 'defaulted';
  investment_date: string;
  maturity_date: string;
  return_date?: string;
  financing_request: {
    id: string;
    requested_amount: number;
    interest_rate: number;
    risk_level: string;
    status: string;
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
  };
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`investment-tabpanel-${index}`}
      aria-labelledby={`investment-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

const InvestmentsPage: React.FC = () => {
  const [investments, setInvestments] = useState<Investment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [tabValue, setTabValue] = useState(0);
  const { user } = useAuth();

  useEffect(() => {
    fetchInvestments();
  }, []);

  const fetchInvestments = async () => {
    try {
      setLoading(true);
      const response = await axios.get('/financing/investments');
      setInvestments(response.data);
    } catch (err: any) {
      setError('Failed to fetch investments');
      console.error('Error fetching investments:', err);
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
      active: 'primary',
      completed: 'success',
      defaulted: 'error'
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

  const calculatePortfolioStats = () => {
    const totalInvested = investments.reduce((sum, inv) => sum + inv.amount, 0);
    const totalExpectedReturn = investments.reduce((sum, inv) => sum + inv.expected_return, 0);
    const totalActualReturn = investments.reduce((sum, inv) => sum + inv.actual_return, 0);
    const activeInvestments = investments.filter(inv => inv.status === 'active').length;
    
    return {
      totalInvested,
      totalExpectedReturn,
      totalActualReturn,
      activeInvestments,
      totalInvestments: investments.length
    };
  };

  const filterInvestmentsByTab = (tabIndex: number) => {
    switch (tabIndex) {
      case 0: return investments; // All
      case 1: return investments.filter(inv => inv.status === 'active');
      case 2: return investments.filter(inv => inv.status === 'completed');
      case 3: return investments.filter(inv => inv.status === 'pending');
      default: return investments;
    }
  };

  const stats = calculatePortfolioStats();
  const filteredInvestments = filterInvestmentsByTab(tabValue);

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
          My Investments
        </Typography>
        <Typography variant="body1" color="textSecondary" sx={{ mb: 4 }}>
          Track your investment portfolio and returns
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
            {error}
          </Alert>
        )}

        {/* Portfolio Statistics */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <AttachMoney sx={{ fontSize: 40, color: 'primary.main', mr: 2 }} />
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
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <TrendingUp sx={{ fontSize: 40, color: 'success.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Expected Returns
                    </Typography>
                    <Typography variant="h5">
                      {formatCurrency(stats.totalExpectedReturn)}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Assessment sx={{ fontSize: 40, color: 'info.main', mr: 2 }} />
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
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <DateRange sx={{ fontSize: 40, color: 'secondary.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Total Investments
                    </Typography>
                    <Typography variant="h5">
                      {stats.totalInvestments}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Investment Tabs */}
        <Paper sx={{ mb: 3 }}>
          <Tabs 
            value={tabValue} 
            onChange={(e, newValue) => setTabValue(newValue)}
            indicatorColor="primary"
            textColor="primary"
          >
            <Tab label={`All (${investments.length})`} />
            <Tab label={`Active (${investments.filter(inv => inv.status === 'active').length})`} />
            <Tab label={`Completed (${investments.filter(inv => inv.status === 'completed').length})`} />
            <Tab label={`Pending (${investments.filter(inv => inv.status === 'pending').length})`} />
          </Tabs>
        </Paper>

        {/* Investments Table */}
        <TabPanel value={tabValue} index={tabValue}>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Invoice</TableCell>
                  <TableCell>Company</TableCell>
                  <TableCell>Amount</TableCell>
                  <TableCell>Expected Return</TableCell>
                  <TableCell>Risk Level</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Investment Date</TableCell>
                  <TableCell>Maturity Date</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredInvestments.map((investment) => (
                  <TableRow key={investment.id}>
                    <TableCell>
                      <Box>
                        <Typography variant="body2" fontWeight="medium">
                          {investment.financing_request.invoice.invoice_number}
                        </Typography>
                        <Typography variant="caption" color="textSecondary">
                          Customer: {investment.financing_request.invoice.customer_name}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Box>
                        <Typography variant="body2">
                          {investment.financing_request.invoice.user.company_name}
                        </Typography>
                        <Typography variant="caption" color="textSecondary">
                          Credit Score: {investment.financing_request.invoice.user.credit_score}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" fontWeight="medium">
                        {formatCurrency(investment.amount)}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" color="success.main" fontWeight="medium">
                        {formatCurrency(investment.expected_return)}
                      </Typography>
                      <Typography variant="caption" color="textSecondary">
                        ({investment.financing_request.interest_rate}%)
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={investment.financing_request.risk_level.toUpperCase()}
                        color={getRiskColor(investment.financing_request.risk_level) as any}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={investment.status.charAt(0).toUpperCase() + investment.status.slice(1)}
                        color={getStatusColor(investment.status) as any}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      {formatDate(investment.investment_date)}
                    </TableCell>
                    <TableCell>
                      {formatDate(investment.maturity_date)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </TabPanel>

        {investments.length === 0 && (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <Typography variant="h6" color="textSecondary" gutterBottom>
              No investments found
            </Typography>
            <Typography variant="body2" color="textSecondary">
              Start investing in the marketplace to see your portfolio here
            </Typography>
          </Box>
        )}
      </Box>
    </Container>
  );
};

export default InvestmentsPage;
