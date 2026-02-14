package com.example.demobank

import android.content.Context
import android.content.Intent
import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.view.Menu
import android.view.MenuItem
import androidx.appcompat.widget.Toolbar
import com.google.android.material.bottomnavigation.BottomNavigationView

class HomeActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_home)

        val toolbar = findViewById<Toolbar>(R.id.toolbar)
        setSupportActionBar(toolbar)
        supportActionBar?.setDisplayHomeAsUpEnabled(true)
        supportActionBar?.setHomeAsUpIndicator(R.drawable.ic_menu)

        val bottomNavigationView = findViewById<BottomNavigationView>(R.id.bottom_navigation)
        bottomNavigationView.setOnItemSelectedListener {
            val sharedPref = getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
            val token = sharedPref.getString("TOKEN", null)
            val username = intent.getStringExtra("USERNAME")

            val bundle = Bundle()
            bundle.putString("TOKEN", token)
            bundle.putString("USERNAME", username)

            val fragment = when (it.itemId) {
                R.id.nav_home -> DashboardFragment()
                R.id.nav_cards -> CardsFragment()
                R.id.nav_transfers -> TransfersFragment()
                R.id.nav_payments -> PaymentsFragment()
                R.id.nav_more -> MoreFragment()
                else -> DashboardFragment()
            }
            fragment.arguments = bundle

            supportFragmentManager.beginTransaction().replace(R.id.fragment_container, fragment).commit()
            true
        }

        // Set the default fragment
        if (savedInstanceState == null) {
            bottomNavigationView.selectedItemId = R.id.nav_home
        }
    }

    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        val sharedPref = getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
        val token = sharedPref.getString("TOKEN", null)
        val username = intent.getStringExtra("USERNAME")

        val bundle = Bundle()
        bundle.putString("TOKEN", token)
        bundle.putString("USERNAME", username)

        when (item.itemId) {
            android.R.id.home -> {
                val fragment = MoreFragment()
                fragment.arguments = bundle
                supportFragmentManager.beginTransaction().replace(R.id.fragment_container, fragment).commit()
                return true
            }
        }
        return super.onOptionsItemSelected(item)
    }
}
