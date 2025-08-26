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
  Receipt,
  AttachMoney,
  TrendingUp,
  Assessment,
  Add,
  Warning,
  CheckCircle
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

interface Invoice {
  id: string;
  invoice_number: string;
  customer_name: string;
  invoice_amount: number;
  due_date: string;
  status: string;
  created_at: string;
}

interface FinancingRequest {
  id: string;
  invoice_id: string;
  requested_amount: number;
  interest_rate: number;
  risk_level: string;
  status: 'pending' | 'approved' | 'rejected' | 'funded';
  created_at: string;
  invoice: {
    invoice_number: string;
    customer_name: string;
    invoice_amount: number;
    due_date: string;
  };
}

interface DashboardStats {
  totalInvoices: number;
  totalInvoiceValue: number;
  activeFinancingRequests: number;
  fundedAmount: number;
  pendingRequests: number;
  approvalRate: number;
}

const SMEDashboard: React.FC = () => {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [financingRequests, setFinancingRequests] = useState<FinancingRequest[]>([]);
  const [stats, setStats] = useState<DashboardStats>({
    totalInvoices: 0,
    totalInvoiceValue: 0,
    activeFinancingRequests: 0,
    fundedAmount: 0,
    pendingRequests: 0,
    approvalRate: 0
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
      const [invoicesResponse, financingResponse] = await Promise.all([
        axios.get('/invoices'),
        axios.get('/financing/requests')
      ]);

      const invoicesData = invoicesResponse.data || [];
      const financingData = financingResponse.data || [];

      setInvoices(invoicesData.slice(0, 5)); // Show recent 5
      setFinancingRequests(financingData.slice(0, 5)); // Show recent 5

      // Calculate stats
      const totalInvoices = invoicesData.length;
      const totalInvoiceValue = invoicesData.reduce((sum: number, inv: Invoice) => sum + inv.invoice_amount, 0);
      const activeFinancingRequests = financingData.filter((req: FinancingRequest) => 
        ['pending', 'approved', 'funded'].includes(req.status)
      ).length;
      const fundedAmount = financingData
        .filter((req: FinancingRequest) => req.status === 'funded')
        .reduce((sum: number, req: FinancingRequest) => sum + req.requested_amount, 0);
      const pendingRequests = financingData.filter((req: FinancingRequest) => req.status === 'pending').length;
      const approvedRequests = financingData.filter((req: FinancingRequest) => req.status === 'approved').length;
      const approvalRate = financingData.length > 0 ? 
        ((approvedRequests + financingData.filter((req: FinancingRequest) => req.status === 'funded').length) / financingData.length) * 100 : 0;

      setStats({
        totalInvoices,
        totalInvoiceValue,
        activeFinancingRequests,
        fundedAmount,
        pendingRequests,
        approvalRate
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
      funded: 'success',
      rejected: 'error',
      paid: 'success',
      overdue: 'error',
      draft: 'default'
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
          SME Dashboard
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
                  <Receipt sx={{ fontSize: 40, color: 'primary.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Total Invoices
                    </Typography>
                    <Typography variant="h5">
                      {stats.totalInvoices}
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
                  <AttachMoney sx={{ fontSize: 40, color: 'success.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Invoice Value
                    </Typography>
                    <Typography variant="h5">
                      {formatCurrency(stats.totalInvoiceValue)}
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
                  <TrendingUp sx={{ fontSize: 40, color: 'info.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Funded Amount
                    </Typography>
                    <Typography variant="h5">
                      {formatCurrency(stats.fundedAmount)}
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
                      Active Requests
                    </Typography>
                    <Typography variant="h5">
                      {stats.activeFinancingRequests}
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
                  <Warning sx={{ fontSize: 40, color: 'warning.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Pending Requests
                    </Typography>
                    <Typography variant="h5">
                      {stats.pendingRequests}
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
                  <CheckCircle sx={{ fontSize: 40, color: 'success.main', mr: 2 }} />
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Approval Rate
                    </Typography>
                    <Typography variant="h5">
                      {stats.approvalRate.toFixed(1)}%
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        <Grid container spacing={3}>
          {/* Recent Invoices */}
          <Grid item xs={12} lg={6}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6" component="h2">
                    Recent Invoices
                  </Typography>
                  <Button
                    startIcon={<Add />}
                    variant="contained"
                    size="small"
                    onClick={() => navigate('/invoices')}
                  >
                    Create Invoice
                  </Button>
                </Box>
                <TableContainer>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Invoice #</TableCell>
                        <TableCell>Customer</TableCell>
                        <TableCell>Amount</TableCell>
                        <TableCell>Status</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {invoices.map((invoice) => (
                        <TableRow key={invoice.id}>
                          <TableCell>
                            <Typography variant="body2" fontWeight="medium">
                              {invoice.invoice_number}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2">
                              {invoice.customer_name}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2">
                              {formatCurrency(invoice.invoice_amount)}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1)}
                              color={getStatusColor(invoice.status) as any}
                              size="small"
                            />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
                {invoices.length === 0 && (
                  <Box sx={{ textAlign: 'center', py: 3 }}>
                    <Typography variant="body2" color="textSecondary">
                      No invoices found. Create your first invoice to get started.
                    </Typography>
                  </Box>
                )}
              </CardContent>
            </Card>
          </Grid>

          {/* Recent Financing Requests */}
          <Grid item xs={12} lg={6}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6" component="h2">
                    Recent Financing Requests
                  </Typography>
                  <Button
                    startIcon={<Add />}
                    variant="contained"
                    size="small"
                    onClick={() => navigate('/invoices')}
                  >
                    Request Financing
                  </Button>
                </Box>
                <TableContainer>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Invoice #</TableCell>
                        <TableCell>Amount</TableCell>
                        <TableCell>Risk</TableCell>
                        <TableCell>Status</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {financingRequests.map((request) => (
                        <TableRow key={request.id}>
                          <TableCell>
                            <Typography variant="body2" fontWeight="medium">
                              {request.invoice.invoice_number}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2">
                              {formatCurrency(request.requested_amount)}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={request.risk_level.toUpperCase()}
                              color={getRiskColor(request.risk_level) as any}
                              size="small"
                            />
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={request.status.charAt(0).toUpperCase() + request.status.slice(1)}
                              color={getStatusColor(request.status) as any}
                              size="small"
                            />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
                {financingRequests.length === 0 && (
                  <Box sx={{ textAlign: 'center', py: 3 }}>
                    <Typography variant="body2" color="textSecondary">
                      No financing requests found. Create an invoice and request financing.
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

export default SMEDashboard;
