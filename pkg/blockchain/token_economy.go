package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// Token represents O2OChat token
type Token struct {
	Name           string
	Symbol         string
	Decimals       int
	TotalSupply    *big.Int
	CirculatingSupply *big.Int
}

// TokenContract represents token smart contract
type TokenContract struct {
	Token           *Token
	Balances        map[string]*big.Int
	Allowances      map[string]map[string]*big.Int
	Transfers       []TransferEvent
	Owner           string
	Paused          bool
	mu              sync.RWMutex
}

// TransferEvent represents a transfer event
type TransferEvent struct {
	From   string
	To     string
	Value  *big.Int
	Time   time.Time
	TxHash string
}

// Stake represents a staking position
type Stake struct {
	User         string
	Amount       *big.Int
	StartTime    time.Time
	EndTime      time.Time
	Rewards      *big.Int
	ClaimedRewards *big.Int
	Status       string // "active", "completed", "withdrawn"
}

// StakingContract represents staking smart contract
type StakingContract struct {
	TokenContract *TokenContract
	Stakes        map[string][]*Stake
	StakingToken  string
	RewardRate    *big.Int // Rewards per second
	TotalStaked   *big.Int
	MinStakingPeriod time.Duration
	mu            sync.RWMutex
}

// GovernanceProposal represents a governance proposal
type GovernanceProposal struct {
	ID          string
	Proposer    string
	Title       string
	Description string
	ForVotes    *big.Int
	AgainstVotes *big.Int
	AbstainVotes *big.Int
	StartTime   time.Time
	EndTime     time.Time
	Status      string // "active", "succeeded", "defeated", "executed"
	Executed    bool
}

// GovernanceContract represents governance smart contract
type GovernanceContract struct {
	TokenContract *TokenContract
	Proposals     map[string]*GovernanceProposal
	Voters        map[string]map[string]string // user -> proposal -> vote
	MinProposalStake *big.Int
	VotingPeriod  time.Duration
	QuorumPercent int
	mu            sync.RWMutex
}

// NewTokenContract creates a new token contract
func NewTokenContract(name, symbol string, totalSupply *big.Int, owner string) *TokenContract {
	return &TokenContract{
		Token: &Token{
			Name:           name,
			Symbol:         symbol,
			Decimals:       18,
			TotalSupply:    totalSupply,
			CirculatingSupply: big.NewInt(0),
		},
		Balances:   make(map[string]*big.Int),
		Allowances: make(map[string]map[string]*big.Int),
		Transfers:  make([]TransferEvent, 0),
		Owner:      owner,
		Paused:     false,
	}
}

// Mint mints new tokens
func (tc *TokenContract) Mint(ctx context.Context, to string, amount *big.Int) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.Paused {
		return fmt.Errorf("contract is paused")
	}

	if to == "" {
		return fmt.Errorf("invalid recipient")
	}

	// Mint tokens
	currentBalance := tc.Balances[to]
	if currentBalance == nil {
		currentBalance = big.NewInt(0)
	}
	tc.Balances[to] = currentBalance.Add(currentBalance, amount)

	// Update circulating supply
	tc.Token.CirculatingSupply = tc.Token.CirculatingSupply.Add(tc.Token.CirculatingSupply, amount)

	// Record transfer event
	tc.Transfers = append(tc.Transfers, TransferEvent{
		From:  "0x0",
		To:    to,
		Value: amount,
		Time:  time.Now(),
	})

	return nil
}

// Transfer transfers tokens
func (tc *TokenContract) Transfer(ctx context.Context, from, to string, amount *big.Int) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.Paused {
		return fmt.Errorf("contract is paused")
	}

	// Check balance
	balance := tc.Balances[from]
	if balance == nil || balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance")
	}

	// Transfer tokens
	tc.Balances[from] = balance.Sub(balance, amount)

	toBalance := tc.Balances[to]
	if toBalance == nil {
		toBalance = big.NewInt(0)
	}
	tc.Balances[to] = toBalance.Add(toBalance, amount)

	// Record transfer event
	tc.Transfers = append(tc.Transfers, TransferEvent{
		From:  from,
		To:    to,
		Value: amount,
		Time:  time.Now(),
	})

	return nil
}

// GetBalance gets token balance
func (tc *TokenContract) GetBalance(address string) *big.Int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	balance := tc.Balances[address]
	if balance == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(balance)
}

// NewStakingContract creates a new staking contract
func NewStakingContract(tokenContract *TokenContract, rewardRate *big.Int) *StakingContract {
	return &StakingContract{
		TokenContract:  tokenContract,
		Stakes:         make(map[string][]*Stake),
		RewardRate:     rewardRate,
		TotalStaked:    big.NewInt(0),
		MinStakingPeriod: 7 * 24 * time.Hour, // 7 days
	}
}

// Stake stakes tokens
func (sc *StakingContract) Stake(ctx context.Context, user string, amount *big.Int) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Transfer tokens to staking contract
	err := sc.TokenContract.Transfer(ctx, user, "staking_contract", amount)
	if err != nil {
		return err
	}

	// Create stake
	stake := &Stake{
		User:         user,
		Amount:       new(big.Int).Set(amount),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(sc.MinStakingPeriod),
		Rewards:      big.NewInt(0),
		ClaimedRewards: big.NewInt(0),
		Status:       "active",
	}

	sc.Stakes[user] = append(sc.Stakes[user], stake)
	sc.TotalStaked = sc.TotalStaked.Add(sc.TotalStaked, amount)

	return nil
}

// ClaimRewards claims staking rewards
func (sc *StakingContract) ClaimRewards(ctx context.Context, user string) (*big.Int, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	stakes := sc.Stakes[user]
	if len(stakes) == 0 {
		return big.NewInt(0), fmt.Errorf("no stakes found")
	}

	totalRewards := big.NewInt(0)

	for _, stake := range stakes {
		if stake.Status != "active" {
			continue
		}

		// Calculate rewards
		elapsed := time.Since(stake.StartTime).Seconds()
		rewards := new(big.Int).Mul(sc.RewardRate, big.NewInt(int64(elapsed)))

		stake.Rewards = rewards
		stake.ClaimedRewards = new(big.Int).Add(stake.ClaimedRewards, rewards)

		totalRewards = totalRewards.Add(totalRewards, rewards)
	}

	// Transfer rewards
	if totalRewards.Cmp(big.NewInt(0)) > 0 {
		err := sc.TokenContract.Transfer(ctx, "reward_pool", user, totalRewards)
		if err != nil {
			return nil, err
		}
	}

	return totalRewards, nil
}

// GetStakingStats gets staking statistics
func (sc *StakingContract) GetStakingStats() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	totalStakers := len(sc.Stakes)
	totalStaked := new(big.Int).Set(sc.TotalStaked)

	return map[string]interface{}{
		"total_stakers":    totalStakers,
		"total_staked":     totalStaked.String(),
		"reward_rate":      sc.RewardRate.String(),
		"min_staking_period": sc.MinStakingPeriod.String(),
	}
}

// NewGovernanceContract creates a new governance contract
func NewGovernanceContract(tokenContract *TokenContract, minStake *big.Int) *GovernanceContract {
	return &GovernanceContract{
		TokenContract:    tokenContract,
		Proposals:        make(map[string]*GovernanceProposal),
		Voters:           make(map[string]map[string]string),
		MinProposalStake: minStake,
		VotingPeriod:     7 * 24 * time.Hour, // 7 days
		QuorumPercent:    10, // 10%
	}
}

// CreateProposal creates a new governance proposal
func (gc *GovernanceContract) CreateProposal(ctx context.Context, proposer, title, description string) (string, error) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	// Check minimum stake
	balance := gc.TokenContract.GetBalance(proposer)
	if balance.Cmp(gc.MinProposalStake) < 0 {
		return "", fmt.Errorf("insufficient stake to create proposal")
	}

	proposalID := fmt.Sprintf("proposal-%d", time.Now().UnixNano())

	proposal := &GovernanceProposal{
		ID:          proposalID,
		Proposer:    proposer,
		Title:       title,
		Description: description,
		ForVotes:    big.NewInt(0),
		AgainstVotes: big.NewInt(0),
		AbstainVotes: big.NewInt(0),
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(gc.VotingPeriod),
		Status:      "active",
		Executed:    false,
	}

	gc.Proposals[proposalID] = proposal
	gc.Voters[proposer] = make(map[string]string)

	return proposalID, nil
}

// Vote votes on a proposal
func (gc *GovernanceContract) Vote(ctx context.Context, voter, proposalID, vote string) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	proposal, exists := gc.Proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found")
	}

	if proposal.Status != "active" {
		return fmt.Errorf("proposal is not active")
	}

	if time.Now().After(proposal.EndTime) {
		return fmt.Errorf("voting period ended")
	}

	// Check voter balance
	balance := gc.TokenContract.GetBalance(voter)
	if balance.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("no voting power")
	}

	// Record vote
	if _, exists := gc.Voters[voter]; !exists {
		gc.Voters[voter] = make(map[string]string)
	}
	gc.Voters[voter][proposalID] = vote

	// Update vote counts
	switch vote {
	case "for":
		proposal.ForVotes = proposal.ForVotes.Add(proposal.ForVotes, balance)
	case "against":
		proposal.AgainstVotes = proposal.AgainstVotes.Add(proposal.AgainstVotes, balance)
	case "abstain":
		proposal.AbstainVotes = proposal.AbstainVotes.Add(proposal.AbstainVotes, balance)
	}

	return nil
}

// ExecuteProposal executes a proposal
func (gc *GovernanceContract) ExecuteProposal(ctx context.Context, proposalID string) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	proposal, exists := gc.Proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found")
	}

	if proposal.Status != "active" {
		return fmt.Errorf("proposal is not active")
	}

	if time.Now().Before(proposal.EndTime) {
		return fmt.Errorf("voting period not ended")
	}

	// Check quorum
	totalVotes := new(big.Int).Add(proposal.ForVotes, proposal.AgainstVotes)
	totalVotes = totalVotes.Add(totalVotes, proposal.AbstainVotes)

	totalSupply := gc.TokenContract.Token.TotalSupply
	quorum := new(big.Int).Mul(totalSupply, big.NewInt(int64(gc.QuorumPercent)))
	quorum = quorum.Div(quorum, big.NewInt(100))

	if totalVotes.Cmp(quorum) < 0 {
		proposal.Status = "defeated"
		return fmt.Errorf("quorum not reached")
	}

	// Check if passed
	if proposal.ForVotes.Cmp(proposal.AgainstVotes) > 0 {
		proposal.Status = "succeeded"
		proposal.Executed = true
	} else {
		proposal.Status = "defeated"
	}

	return nil
}

// GetGovernanceStats gets governance statistics
func (gc *GovernanceContract) GetGovernanceStats() map[string]interface{} {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	totalProposals := len(gc.Proposals)
	activeProposals := 0
	succeededProposals := 0
	defeatedProposals := 0

	for _, proposal := range gc.Proposals {
		switch proposal.Status {
		case "active":
			activeProposals++
		case "succeeded":
			succeededProposals++
		case "defeated":
			defeatedProposals++
		}
	}

	return map[string]interface{}{
		"total_proposals":     totalProposals,
		"active_proposals":    activeProposals,
		"succeeded_proposals": succeededProposals,
		"defeated_proposals":  defeatedProposals,
		"quorum_percent":      gc.QuorumPercent,
		"voting_period":       gc.VotingPeriod.String(),
	}
}
