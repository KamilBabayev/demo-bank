package com.example.demobank

import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.widget.TextView

class NotificationDetailActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_notification_detail)

        val notification = intent.getParcelableExtra<Notification>("NOTIFICATION")

        if (notification != null) {
            val titleTextView = findViewById<TextView>(R.id.notification_title)
            val contentTextView = findViewById<TextView>(R.id.notification_content)
            val dateTextView = findViewById<TextView>(R.id.notification_date)
            val metadataTextView = findViewById<TextView>(R.id.notification_metadata)

            titleTextView.text = notification.title
            contentTextView.text = notification.content
            dateTextView.text = notification.createdAt

            val metadata = notification.metadata
            var metadataText = ""
            if (metadata.payment_id != null) {
                metadataText += "Payment ID: ${metadata.payment_id}\n"
            }
            if (metadata.transfer_id != null) {
                metadataText += "Transfer ID: ${metadata.transfer_id}\n"
            }
            if (metadata.reference_id != null) {
                metadataText += "Reference ID: ${metadata.reference_id}\n"
            }
            if (metadata.failure_reason != null) {
                metadataText += "Failure Reason: ${metadata.failure_reason}"
            }
            metadataTextView.text = metadataText
        }
    }
}
