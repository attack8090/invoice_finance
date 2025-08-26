import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  Card,
  CardContent,
  CardActions,
  Button,
  Grid,
  Chip,
  Alert,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Divider,
  Avatar,
  LinearProgress,
  IconButton,
  Tooltip
} from '@mui/material';
import {
  TrendingUp,
  Security,
  AccessTime,
  AttachMoney,
  Business,
  Info as InfoIcon,
  CheckCircle as VerifiedIcon
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import axios from 'axios';

interface FinancingRequest {
  id: string;
  invoice_id: string;
  requested_amount: number;
  interest_rate: number;
  financing_fee: number;
  net_amount: number;
  status: string;
  description: string;
  expected_return: number;
  risk_level: 'low' | 'medium' | 'high';
  created_at: string;
  invoice: {
    id: string;
    invoice_number: string;
    customer_name: string;
    invoice_amount: number;
    due_date: string;
    issue_date: string;
    verification_status: string;
    document_url?: string;
  };
  user: {
    id: string;
    company_name: string;
    first_name: string;
    last_name: string;
    credit_score: number;
    is_verified: boolean;
  };
  investments?: any[];
}

const MarketplacePage: React.FC = () => {
  const [opportunities, setOpportunities] = useState<FinancingRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedOpportunity, setSelectedOpportunity] = useState<FinancingRequest | null>(null);
  const [investmentDialogOpen, setInvestmentDialogOpen] = useState(false);
  const [investmentAmount, setInvestmentAmount] = useState('');
  const [investing, setInvesting] = useState(false);
  const { user } = useAuth();

  useEffect(() => {
    fetchOpportunities();
  }, []);

  const fetchOpportunities = async () => {
    try {
      setLoading(true);
      const response = await axios.get('/financing/opportunities?limit=50');
      setOpportunities(response.data);
    } catch (err: any) {
      setError('Failed to fetch investment opportunities');
      console.error('Error fetching opportunities:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleInvestClick = (opportunity: FinancingRequest) => {
    setSelectedOpportunity(opportunity);
    setInvestmentAmount('');
    setInvestmentDialogOpen(true);
  };

  const handleInvestmentSubmit = async () => {
    if (!selectedOpportunity || !investmentAmount) return;

    setInvesting(true);
    try {
      await axios.post('/financing/invest', {
        financing_request_id: selectedOpportunity.id,
        amount: parseFloat(investmentAmount)
      });
      
      setInvestmentDialogOpen(false);
      setSelectedOpportunity(null);
      setInvestmentAmount('');
      fetchOpportunities(); // Refresh data
      
      // Show success message
      setError('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create investment');
    } finally {
      setInvesting(false);
    }
  };

  const getRiskColor = (riskLevel: string) => {
    switch (riskLevel) {
      case 'low': return 'success';
      case 'medium': return 'warning';
      case 'high': return 'error';
      default: return 'default';
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

  const calculateDaysToMaturity = (dueDate: string) => {
    const due = new Date(dueDate);
    const now = new Date();
    const diffTime = due.getTime() - now.getTime();
    return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  };

  const calculateExpectedReturn = (amount: number, interestRate: number) => {
    return amount * (interestRate / 100);
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
        <Box sx={{ mb: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Investment Marketplace
          </Typography>
          <Typography variant="body1" color="textSecondary">
            Discover verified investment opportunities in invoice financing
          </Typography>
        </Box>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
            {error}
          </Alert>
        )}

        {/* Statistics Summary */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card sx={{ textAlign: 'center', p: 2 }}>
              <Typography variant="h4" color="primary">
                {opportunities.length}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Available Opportunities
              </Typography>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card sx={{ textAlign: 'center', p: 2 }}>
              <Typography variant="h4" color="primary">
                {formatCurrency(
                  opportunities.reduce((sum, opp) => sum + opp.requested_amount, 0)
                )}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Total Investment Volume
              </Typography>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card sx={{ textAlign: 'center', p: 2 }}>
              <Typography variant="h4" color="primary">
                {opportunities.length > 0 
                  ? (opportunities.reduce((sum, opp) => sum + opp.interest_rate, 0) / opportunities.length).toFixed(1)
                  : '0.0'}%
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Average Return
              </Typography>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card sx={{ textAlign: 'center', p: 2 }}>
              <Typography variant="h4" color="primary">
                {opportunities.filter(opp => opp.risk_level === 'low').length}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Low Risk Opportunities
              </Typography>
            </Card>
          </Grid>
        </Grid>

        {/* Investment Opportunities */}
        <Grid container spacing={3}>
          {opportunities.map((opportunity) => {
            const daysToMaturity = calculateDaysToMaturity(opportunity.invoice.due_date);
            const fundingProgress = opportunity.investments ? 
              (opportunity.investments.reduce((sum: number, inv: any) => sum + inv.amount, 0) / opportunity.requested_amount) * 100 : 0;
            
            return (
              <Grid item xs={12} md={6} lg={4} key={opportunity.id}>
                <Card 
                  sx={{ 
                    height: '100%', 
                    display: 'flex', 
                    flexDirection: 'column',
                    '&:hover': {
                      boxShadow: 4,
                      transform: 'translateY(-2px)',
                      transition: 'all 0.2s ease-in-out'
                    }
                  }}
                >
                  <CardContent sx={{ flexGrow: 1 }}>
                    {/* Header */}
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                      <Typography variant="h6" component="h2">
                        {opportunity.invoice.invoice_number}
                      </Typography>
                      <Chip 
                        label={opportunity.risk_level.toUpperCase()} 
                        color={getRiskColor(opportunity.risk_level) as any}
                        size="small"
                      />
                    </Box>

                    {/* Company Info */}
                    <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                      <Avatar sx={{ mr: 2, bgcolor: 'primary.main' }}>
                        <Business />
                      </Avatar>
                      <Box>
                        <Typography variant="body2" fontWeight="medium">
                          {opportunity.user.company_name}
                        </Typography>
                        <Box sx={{ display: 'flex', alignItems: 'center' }}>
                          <Typography variant="caption" color="textSecondary" sx={{ mr: 1 }}>
                            Credit Score: {opportunity.user.credit_score}
                          </Typography>
                          {opportunity.user.is_verified && (
                            <Tooltip title="Verified Company">
                              <VerifiedIcon sx={{ fontSize: 16, color: 'success.main' }} />
                            </Tooltip>
                          )}
                        </Box>
                      </Box>
                    </Box>

                    {/* Investment Details */}
                    <Box sx={{ mb: 2 }}>
                      <Grid container spacing={2}>
                        <Grid item xs={6}>
                          <Typography variant="body2" color="textSecondary">Amount</Typography>
                          <Typography variant="h6">{formatCurrency(opportunity.requested_amount)}</Typography>
                        </Grid>
                        <Grid item xs={6}>
                          <Typography variant="body2" color="textSecondary">Return Rate</Typography>
                          <Typography variant="h6" color="success.main">{opportunity.interest_rate}%</Typography>
                        </Grid>
                        <Grid item xs={6}>
                          <Typography variant="body2" color="textSecondary">Customer</Typography>
                          <Typography variant="body2">{opportunity.invoice.customer_name}</Typography>
                        </Grid>
                        <Grid item xs={6}>
                          <Typography variant="body2" color="textSecondary">Maturity</Typography>
                          <Typography variant="body2">{daysToMaturity} days</Typography>
                        </Grid>
                      </Grid>
                    </Box>

                    {/* Funding Progress */}
                    <Box sx={{ mb: 2 }}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                        <Typography variant="body2" color="textSecondary">Funding Progress</Typography>
                        <Typography variant="body2" color="textSecondary">
                          {Math.min(fundingProgress, 100).toFixed(0)}%
                        </Typography>
                      </Box>
                      <LinearProgress 
                        variant="determinate" 
                        value={Math.min(fundingProgress, 100)} 
                        sx={{ height: 8, borderRadius: 4 }}
                      />
                    </Box>

                    {/* Description */}
                    {opportunity.description && (
                      <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
                        {opportunity.description.length > 100 
                          ? `${opportunity.description.substring(0, 100)}...`
                          : opportunity.description
                        }
                      </Typography>
                    )}
                  </CardContent>

                  <CardActions sx={{ p: 2, pt: 0 }}>
                    <Button 
                      fullWidth
                      variant="contained" 
                      onClick={() => handleInvestClick(opportunity)}
                      disabled={user?.role !== 'investor'}
                    >
                      Invest Now
                    </Button>
                    <Tooltip title="View Details">
                      <IconButton size="small">
                        <InfoIcon />
                      </IconButton>
                    </Tooltip>
                  </CardActions>
                </Card>
              </Grid>
            );
          })}
        </Grid>

        {opportunities.length === 0 && (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <Typography variant="h6" color="textSecondary" gutterBottom>
              No investment opportunities available
            </Typography>
            <Typography variant="body2" color="textSecondary">
              Check back later for new opportunities
            </Typography>
          </Box>
        )}
      </Box>

      {/* Investment Dialog */}
      <Dialog 
        open={investmentDialogOpen} 
        onClose={() => setInvestmentDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          Invest in {selectedOpportunity?.invoice.invoice_number}
        </DialogTitle>
        <DialogContent>
          {selectedOpportunity && (
            <Box>
              <Typography variant="body2" color="textSecondary" sx={{ mb: 3 }}>
                You are about to invest in an invoice from {selectedOpportunity.user.company_name}
              </Typography>
              
              <Grid container spacing={2} sx={{ mb: 3 }}>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">Invoice Amount</Typography>
                  <Typography variant="h6">
                    {formatCurrency(selectedOpportunity.invoice.invoice_amount)}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">Return Rate</Typography>
                  <Typography variant="h6" color="success.main">
                    {selectedOpportunity.interest_rate}%
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">Risk Level</Typography>
                  <Chip 
                    label={selectedOpportunity.risk_level.toUpperCase()} 
                    color={getRiskColor(selectedOpportunity.risk_level) as any}
                    size="small"
                  />
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">Due Date</Typography>
                  <Typography variant="body2">
                    {formatDate(selectedOpportunity.invoice.due_date)}
                  </Typography>
                </Grid>
              </Grid>
              
              <Divider sx={{ mb: 3 }} />
              
              <TextField
                fullWidth
                label="Investment Amount"
                type="number"
                value={investmentAmount}
                onChange={(e) => setInvestmentAmount(e.target.value)}
                inputProps={{ 
                  min: 100, 
                  max: selectedOpportunity.requested_amount,
                  step: 100
                }}
                helperText={`Min: $100, Max: ${formatCurrency(selectedOpportunity.requested_amount)}`}
                sx={{ mb: 2 }}
              />
              
              {investmentAmount && (
                <Box sx={{ p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
                  <Typography variant="body2" color="textSecondary">Expected Return</Typography>
                  <Typography variant="h6" color="success.main">
                    {formatCurrency(
                      calculateExpectedReturn(parseFloat(investmentAmount), selectedOpportunity.interest_rate)
                    )}
                  </Typography>
                </Box>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setInvestmentDialogOpen(false)}>Cancel</Button>
          <Button 
            onClick={handleInvestmentSubmit} 
            variant="contained"
            disabled={!investmentAmount || parseFloat(investmentAmount) < 100 || investing}
          >
            {investing ? <CircularProgress size={20} /> : 'Confirm Investment'}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default MarketplacePage;
