import type {
  BaseERC20,
  ERC20,
  erc20Abi,
  ERC20Amount,
  ERC20Permit,
  ERC20PermitData,
  PermitType,
} from "reverse-mirage";
import type {
  Account,
  Address,
  Chain,
  ContractFunctionArgs,
  Hash,
  Hex,
  ReadContractParameters,
  SignTypedDataParameters,
  SimulateContractParameters,
  SimulateContractReturnType,
} from "viem";

export type GetERC20Parameters = Omit<
  ReadContractParameters<typeof erc20Abi, "name">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: Pick<BaseERC20, "address" | "chainID"> &
    Partial<Pick<BaseERC20, "blockCreated">>;
};
export type GetERC20ReturnType = ERC20;

export type GetERC20AllowanceParameters<TERC20 extends BaseERC20> = Omit<
  ReadContractParameters<typeof erc20Abi, "allowance">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: TERC20;
  owner: Address;
  spender: Address;
};
export type GetERC20AllowanceReturnType<TERC20 extends BaseERC20> =
  ERC20Amount<TERC20>;

export type GetERC20BalanceOfParameters<TERC20 extends BaseERC20> = Omit<
  ReadContractParameters<typeof erc20Abi, "balanceOf">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: TERC20;
  address: Address;
};
export type GetERC20BalanceOfReturnType<TERC20 extends BaseERC20> =
  ERC20Amount<TERC20>;

export type GetERC20DecimalsParameters = Omit<
  ReadContractParameters<typeof erc20Abi, "decimals">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: Pick<BaseERC20, "address">;
};
export type GetERC20DecimalsReturnType = number;

export type GetERC20DomainSeparatorParameters = Omit<
  ReadContractParameters<typeof erc20Abi, "DOMAIN_SEPARATOR">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: Pick<BaseERC20, "address">;
};
export type GetERC20DomainSeparatorReturnType = Hex;

export type GetERC20NameParameters = Omit<
  ReadContractParameters<typeof erc20Abi, "name">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: Pick<BaseERC20, "address">;
};
export type GetERC20NameReturnType = string;

export type GetERC20PermitParameters = Omit<
  ReadContractParameters<typeof erc20Abi, "name">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: Pick<ERC20Permit, "address" | "chainID"> &
    Partial<Pick<ERC20Permit, "version" | "blockCreated">>;
};
export type GetERC20PermitReturnType = ERC20Permit;

export type GetERC20PermitDataParameters<TERC20 extends ERC20Permit> = Omit<
  ReadContractParameters<typeof erc20Abi, "nonces">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: TERC20;
  address: Address;
};
export type GetERC20PermitDataReturnType<TERC20 extends ERC20Permit> =
  ERC20PermitData<TERC20>;

export type GetERC20PermitNonceParameters = Omit<
  ReadContractParameters<typeof erc20Abi, "nonces">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: ERC20Permit;
  address: Address;
};
export type GetERC20PermitNonceReturnType = bigint;

export type GetERC20SymbolParameters = Omit<
  ReadContractParameters<typeof erc20Abi, "symbol">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: Pick<BaseERC20, "address">;
};
export type GetERC20SymbolReturnType = string;

export type GetERC20TotalSupplyParameters<TERC20 extends BaseERC20> = Omit<
  ReadContractParameters<typeof erc20Abi, "totalSupply">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: TERC20;
};
export type GetERC20TotalSupplyReturnType<TERC20 extends BaseERC20> =
  ERC20Amount<TERC20>;

export type GetIsERC20PermitParameters = Omit<
  ReadContractParameters<typeof erc20Abi, "name">,
  "address" | "abi" | "functionName" | "args"
> & {
  erc20: Pick<BaseERC20, "address" | "chainID"> &
    Partial<Pick<BaseERC20, "blockCreated">> &
    Partial<Pick<ERC20Permit, "version">>;
};
export type GetIsERC20PermitReturnType = ERC20 | ERC20Permit;
export type SignERC20PermitParameters<TAccount extends Account | undefined> =
  Pick<
    SignTypedDataParameters<typeof PermitType, "Permit", TAccount>,
    "account"
  > & {
    permitData: ERC20PermitData<ERC20Permit>;
    spender: Address;
    deadline: bigint;
  };

export type ERC20ApproveParameters = {
  amount: ERC20Amount<BaseERC20>;
  spender: Address;
};
export type SimulateERC20ApproveParameters<
  chain extends Chain | undefined = Chain | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  accountOverride extends Account | Address | undefined =
    | Account
    | Address
    | undefined,
> = Omit<
  SimulateContractParameters<
    typeof erc20Abi,
    "approve",
    ContractFunctionArgs<typeof erc20Abi, "nonpayable" | "payable", "approve">,
    chain,
    chainOverride,
    accountOverride
  >,
  "args" | "address" | "abi" | "functionName"
> & {
  args: ERC20ApproveParameters;
};
export type SimulateERC20ApproveReturnType<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  accountOverride extends Account | Address | undefined =
    | Account
    | Address
    | undefined,
> = SimulateContractReturnType<
  typeof erc20Abi,
  "approve",
  ContractFunctionArgs<typeof erc20Abi, "nonpayable" | "payable", "approve">,
  chain,
  account,
  chainOverride,
  accountOverride
>;
export type ERC20PermitParameters = {
  signature: Hex;
  owner: Address;
  spender: Address;
  permitData: ERC20PermitData<ERC20Permit>;
  deadline: bigint;
};
export type SimulateERC20PermitParameters<
  chain extends Chain | undefined = Chain | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  accountOverride extends Account | Address | undefined =
    | Account
    | Address
    | undefined,
> = Omit<
  SimulateContractParameters<
    typeof erc20Abi,
    "permit",
    ContractFunctionArgs<typeof erc20Abi, "nonpayable" | "payable", "permit">,
    chain,
    chainOverride,
    accountOverride
  >,
  "args" | "address" | "abi" | "functionName"
> & {
  args: ERC20PermitParameters;
};
export type SimulateERC20PermitReturnType<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  accountOverride extends Account | Address | undefined =
    | Account
    | Address
    | undefined,
> = SimulateContractReturnType<
  typeof erc20Abi,
  "permit",
  ContractFunctionArgs<typeof erc20Abi, "nonpayable" | "payable", "permit">,
  chain,
  account,
  chainOverride,
  accountOverride
>;

export type ERC20TransferParameters = {
  amount: ERC20Amount<BaseERC20>;
  to: Address;
};
export type SimulateERC20TransferParameters<
  chain extends Chain | undefined = Chain | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  accountOverride extends Account | Address | undefined =
    | Account
    | Address
    | undefined,
> = Omit<
  SimulateContractParameters<
    typeof erc20Abi,
    "transfer",
    ContractFunctionArgs<typeof erc20Abi, "nonpayable" | "payable", "transfer">,
    chain,
    chainOverride,
    accountOverride
  >,
  "args" | "address" | "abi" | "functionName"
> & {
  args: ERC20TransferParameters;
};
export type SimulateERC20TransferReturnType<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  accountOverride extends Account | Address | undefined =
    | Account
    | Address
    | undefined,
> = SimulateContractReturnType<
  typeof erc20Abi,
  "transfer",
  ContractFunctionArgs<typeof erc20Abi, "nonpayable" | "payable", "transfer">,
  chain,
  account,
  chainOverride,
  accountOverride
>;

export type ERC20TransferFromParameters = {
  amount: ERC20Amount<BaseERC20>;
  from: Address;
  to: Address;
};
export type SimulateERC20TransferFromParameters<
  chain extends Chain | undefined = Chain | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  accountOverride extends Account | Address | undefined =
    | Account
    | Address
    | undefined,
> = Omit<
  SimulateContractParameters<
    typeof erc20Abi,
    "transferFrom",
    ContractFunctionArgs<
      typeof erc20Abi,
      "nonpayable" | "payable",
      "transferFrom"
    >,
    chain,
    chainOverride,
    accountOverride
  >,
  "args" | "address" | "abi" | "functionName"
> & {
  args: ERC20TransferFromParameters;
};
export type SimulateERC20TransferFromReturnType<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  accountOverride extends Account | Address | undefined =
    | Account
    | Address
    | undefined,
> = SimulateContractReturnType<
  typeof erc20Abi,
  "transferFrom",
  ContractFunctionArgs<
    typeof erc20Abi,
    "nonpayable" | "payable",
    "transferFrom"
  >,
  chain,
  account,
  chainOverride,
  accountOverride
>;

export type PublicActionsERC20<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = {
  getERC20: (args: GetERC20Parameters) => Promise<GetERC20ReturnType>;
  getERC20Allowance: <erc20 extends BaseERC20>(
    args: GetERC20AllowanceParameters<erc20>,
  ) => Promise<GetERC20AllowanceReturnType<erc20>>;
  getERC20BalanceOf: <erc20 extends BaseERC20>(
    args: GetERC20BalanceOfParameters<erc20>,
  ) => Promise<GetERC20BalanceOfReturnType<erc20>>;

  getERC20Decimals: (
    args: GetERC20DecimalsParameters,
  ) => Promise<GetERC20DecimalsReturnType>;

  getERC20DomainSeparator: (
    args: GetERC20DomainSeparatorParameters,
  ) => Promise<GetERC20DomainSeparatorReturnType>;

  getERC20Name: (
    args: GetERC20NameParameters,
  ) => Promise<GetERC20NameReturnType>;

  getERC20Permit: (
    args: GetERC20PermitParameters,
  ) => Promise<GetERC20PermitReturnType>;

  getERC20PermitData: <erc20 extends ERC20Permit>(
    args: GetERC20PermitDataParameters<erc20>,
  ) => Promise<GetERC20PermitDataReturnType<erc20>>;

  getERC20PermitNonce: (
    args: GetERC20PermitNonceParameters,
  ) => Promise<GetERC20PermitNonceReturnType>;

  getERC20Symbol: (
    args: GetERC20SymbolParameters,
  ) => Promise<GetERC20SymbolReturnType>;

  getERC20TotalSupply: <TERC20 extends BaseERC20>(
    args: GetERC20TotalSupplyParameters<TERC20>,
  ) => Promise<GetERC20TotalSupplyReturnType<TERC20>>;

  getIsERC20Permit: (
    args: GetIsERC20PermitParameters,
  ) => Promise<GetIsERC20PermitReturnType>;
};

export type WalletActionsERC20<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = {
  signERC20Permit: (args: SignERC20PermitParameters<account>) => Promise<Hash>;

  simulateERC20Approve: <
    chainOverride extends Chain | undefined = undefined,
    accountOverride extends Address | Account | undefined = undefined,
  >(
    args: SimulateERC20ApproveParameters<chain, chainOverride, accountOverride>,
  ) => Promise<
    SimulateERC20ApproveReturnType<
      chain,
      account,
      chainOverride,
      accountOverride
    >
  >;

  simulateERC20Permit: <
    chainOverride extends Chain | undefined = undefined,
    accountOverride extends Address | Account | undefined = undefined,
  >(
    args: SimulateERC20PermitParameters<chain, chainOverride, accountOverride>,
  ) => Promise<
    SimulateERC20PermitReturnType<
      chain,
      account,
      chainOverride,
      accountOverride
    >
  >;

  simulateERC20Transfer: <
    chainOverride extends Chain | undefined = undefined,
    accountOverride extends Address | Account | undefined = undefined,
  >(
    args: SimulateERC20TransferParameters<
      chain,
      chainOverride,
      accountOverride
    >,
  ) => Promise<
    SimulateERC20TransferReturnType<
      chain,
      account,
      chainOverride,
      accountOverride
    >
  >;

  simulateERC20TransferFrom: <
    chainOverride extends Chain | undefined = undefined,
    accountOverride extends Address | Account | undefined = undefined,
  >(
    args: SimulateERC20TransferFromParameters<
      chain,
      chainOverride,
      accountOverride
    >,
  ) => Promise<
    SimulateERC20TransferFromReturnType<
      chain,
      account,
      chainOverride,
      accountOverride
    >
  >;
};
