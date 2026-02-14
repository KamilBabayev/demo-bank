package com.example.demobank

data class AccountResponse(val accounts: List<Account>, val total: Int)

data class CardResponse(val cards: List<Card>, val total: Int)

data class PaymentResponse(val payments: List<Payment>, val total: Int)
