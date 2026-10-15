
export enum TransactionEditPageType {
    Transaction = 'transaction',
    Template = 'template'
}

export enum TransactionEditPageMode {
    Add = 'add',
    Edit = 'edit',
    View = 'view'
}

export enum GeoLocationStatus {
    Getting = 'getting',
    Success = 'success',
    Error = 'error'
}

export enum AfterSaveAction {
    GoBack = 'goBack',
    StayWithNewTransaction = 'stayWithNewTransaction',
    StayWithCurrentTransaction = 'stayWithCurrentTransaction'
}
